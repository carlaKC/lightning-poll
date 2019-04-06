package votes

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"lightning-poll/lnd"
	"log"

	votes_db "lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
)

type Backends interface {
	GetDB() *sql.DB
	GetLND() lnd.Client
}

// Create initiates the process of voting for an option. It queries LND for
// an invoice, saved it in the votes DB and returns it to the user.
func Create(ctx context.Context, b Backends, pollID, optionID, sats, expiry int64) (int64, string, error) {
	inv, err := b.GetLND().AddInvoice(ctx, sats, expiry, fmt.Sprintf("poll: %v, option: %v", pollID, optionID))
	if err != nil {
		return 0, "", err
	}

	id, err := votes_db.Create(ctx, b.GetDB(), pollID, optionID, expiry,
		inv.PaymentRequest, hex.EncodeToString(inv.RHash))
	if err != nil {
		return 0, "", err
	}

	log.Printf("votes/ops: Created vote: %v", id)

	return id, inv.PaymentRequest, nil
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
		voteList = append(voteList, &Vote{ID: vote.ID, OptionID: vote.OptionID})
	}

	return voteList, nil
}
