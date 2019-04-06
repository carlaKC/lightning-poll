package votes

import (
	"context"
	"database/sql"
	"log"

	votes_db "lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
)

type Backends interface {
	GetDB() *sql.DB
	//GetLND() TODO(carla): add LND
}

// Create initiates the process of voting for an option. It queries LND for
// an invoice, saved it in the votes DB and returns it to the user.
func Create(ctx context.Context, b Backends, pollID, optionID, sats, expiry int64) (int64, string, error) {
	// TODO(carla): query LND for invoice
	payReq := "example pay req"

	id, err := votes_db.Create(ctx, b.GetDB(), pollID, optionID, expiry, payReq)
	if err != nil {
		return 0, "", err
	}

	log.Printf("votes/ops: Created vote: %v", id)

	return id, payReq, nil
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
