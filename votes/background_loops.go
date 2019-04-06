package votes

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	votes_db "lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
)

func Start(b Backends) {
	go expireVotesForever(b)
	go settleInvoicesForever(b)
}

func expireVotesForever(b Backends) {
	for {
		if err := expireVotes(context.Background(), b); err != nil {
			log.Printf("votes/ops: expireVotesForever error: %v", err)
		}
		time.Sleep(time.Minute * 5)
	}
}

func expireVotes(ctx context.Context, b Backends) error {
	expired, err := votes_db.ListExpired(ctx, b.GetDB())
	if err != nil {
		return err
	}

	if len(expired) == 0 {
		return nil
	}

	for _, exp := range expired {
		// if the invoice has been settled, update vote to paid
		inv, err := b.GetLND().LookupInvoice(ctx, exp.PayReq)
		if err != nil {
			return err
		}

		if inv.SettleIndex > 0 {
			if err := settleInvoice(ctx, b, hex.EncodeToString(inv.RHash),
				inv.AmtPaidSat, inv.SettleIndex); err != nil {
				return err
			}
		}

		// if the invoice has not been paid, expire it
		if err := votes_db.UpdateStatus(ctx, b.GetDB(), exp.ID, types.VoteStatusCreated,
			types.VoteStatusExpired); err != nil {
			return err
		}
	}
	return nil
}

func settleInvoicesForever(b Backends) {
	for {
		if err := settleInvoices(b); err != nil {
			log.Printf("votes/ops: settleInvoicesForever error: %v", err)
		}
		time.Sleep(time.Minute)
	}
}

func settleInvoices(b Backends) error {
	ctx := context.Background()

	index, err := votes_db.GetLatestSettleIndex(ctx, b.GetDB())
	if err != nil {
		return err
	}

	log.Printf("votes/ops: settleInvoices resuming at %v", index)

	cl, err := b.GetLND().SubscribeInvoices(ctx, index)
	if err != nil {
		return err
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		inv, err := cl.Recv()
		if err != nil {
			return err
		}

		if inv.SettleIndex <= 0 {
			log.Printf("votes/ops: settleInvoices stream received a non-settled invoice")
			continue
		}

		if err := settleInvoice(ctx, b, hex.EncodeToString(inv.RHash), inv.AmtPaidSat, inv.SettleIndex); err != nil {
			return err
		}
	}
}

func settleInvoice(ctx context.Context, b Backends, payHash string, settledAmount int64, settleIndex uint64) error {
	vote, err := votes_db.LookupByHash(ctx, b.GetDB(), payHash)
	if err != nil {
		return err
	}

	return votes_db.Settle(ctx, b.GetDB(), vote.ID, settledAmount, settleIndex)
}
