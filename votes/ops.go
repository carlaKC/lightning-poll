package votes

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"lightning-poll/lnd"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"

	ext_types "lightning-poll/types"
	votes_db "lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
)

type Backends interface {
	GetDB() *sql.DB
	GetLND() lnd.Client
}

// Create initiates the process of voting for an option. It queries LND for
// an invoice, saved it in the votes DB and returns it to the user.
func Create(ctx context.Context, b Backends, pollID, optionID, sats, expiry int64) (int64, error) {
	resp, err := b.GetLND().AddHoldInvoice(ctx, sats, expiry, fmt.Sprintf("poll: %v, option: %v", pollID, optionID))
	if err != nil {
		return 0, err
	}

	id, err := votes_db.Create(ctx, b.GetDB(), pollID, optionID, expiry,
		resp.PayReq, resp.PayHash, resp.Preimage)
	if err != nil {
		return 0, err
	}

	log.Printf("votes/ops: Created vote: %v", id)
	go subscribeIndividualInvoice(context.Background(), b, id, resp.PayHash)

	return id, nil
}

func Lookup(ctx context.Context, b Backends, id int64) (*Vote, error) {
	vote, err := votes_db.Lookup(ctx, b.GetDB(), id)
	if err != nil {
		return nil, err
	}

	return &Vote{
		ID:       vote.ID,
		PollID:   vote.PollID,
		OptionID: vote.OptionID,
		Amount:   vote.SettleAmount,
		PayReq:   vote.PayReq,
	}, nil
}

func subscribeIndividualInvoice(ctx context.Context, b Backends, id int64, payHash string) {
	cl, err := b.GetLND().SubscribeInvoice(ctx, id, payHash)
	if err != nil {
		log.Printf("subscribeIndividualInvoice error:%v", err)
		return
	}

	maxIterations := 5
	count := 0
	for {
		// just a sanity check so that these goroutines don't spiral off into infinity
		count++
		if count == maxIterations {
			log.Printf("subscribeIndividualInvoice breaking out to prevent leak:%v", err)
		}

		if ctx.Err() != nil {
			log.Printf("subscribeIndividualInvoice error:%v", err)
			return
		}

		inv, err := cl.Recv()
		if err != nil {
			log.Printf("subscribeIndividualInvoice error:%v", err)
			return
		}

		if inv.State != lnrpc.Invoice_ACCEPTED {
			log.Printf("votes/ops: settleInvoices stream received a non-settled invoice")
			continue
		}

		if err := markInvoicePaid(ctx, b, hex.EncodeToString(inv.RHash), inv.AmtPaidSat, inv.SettleIndex); err != nil {
			log.Printf("subscribeIndividualInvoice error:%v", err)
			return
		}
		log.Printf("votes/ops: marked invoice %v as paid", id)
		return
	}
}

// GetResults returns a map of options IDs to vote counts.
// Note that only paid votes are included.
func GetResults(ctx context.Context, b Backends, pollID int64) (map[int64]int64, error) {
	v := make(map[int64]int64)

	votes, err := votes_db.ListByPollAndStatus(ctx, b.GetDB(), pollID, types.VoteStatusPaid)
	if err != nil {
		return v, err
	}

	for _, vote := range votes {
		v[vote.OptionID] = v[vote.OptionID] + 1
	}

	return v, nil
}

func GetVotes(ctx context.Context, b Backends, pollID int64) ([]*Vote, error) {
	votes, err := votes_db.ListByPollAndStatus(ctx, b.GetDB(), pollID, types.VoteStatusPaid)
	if err != nil {
		return nil, err
	}

	var voteList []*Vote
	for _, vote := range votes {
		voteList = append(voteList, &Vote{
			ID:       vote.ID,
			OptionID: vote.OptionID,
			Preimage: vote.Preimage,
			Hash:     vote.PayHash,
			Amount:   vote.SettleAmount,
		})
	}

	return voteList, nil
}

func ReleaseVotesForPoll(ctx context.Context, b Backends, pollID int64, shouldRepay ext_types.RepaySchemeFunc) (int64, error) {
	results, err := GetResults(ctx, b, pollID)
	if err != nil {
		return 0, err
	}

	votes, err := GetVotes(ctx, b, pollID)
	if err != nil {
		return 0, err
	}

	var amount int64 = 0

	for _, vote := range votes {
		if shouldRepay(results, vote.OptionID) {
			if err := releaseVote(ctx, b, vote.ID, vote.Hash); err != nil {
				return 0, err
			}
		} else {
			amount = amount + vote.Amount
			if err := settleVote(ctx, b, vote.ID, vote.Preimage); err != nil {
				return 0, err
			}
		}
	}

	return amount, nil
}

func releaseVote(ctx context.Context, b Backends, id int64, hash string) error {
	if err := b.GetLND().CancelHoldInvoice(ctx, hash); err != nil {
		return err
	}
	if err := votes_db.UpdateStatus(ctx, b.GetDB(), id, types.VoteStatusPaid,
		types.VoteStatusReturned); err != nil {
		return err
	}

	return nil
}

func settleVote(ctx context.Context, b Backends, id int64, preimage []byte) error {
	if err := b.GetLND().SettleHoldInvoice(ctx, preimage); err != nil {
		return err
	}
	if err := votes_db.UpdateStatus(ctx, b.GetDB(), id, types.VoteStatusPaid,
		types.VoteStatusSettled); err != nil {
		return err
	}

	return nil
}
