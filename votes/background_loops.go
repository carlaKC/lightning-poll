package votes

import (
	"context"
	"log"
	"time"

	votes_db "lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
)

func Start(b Backends) {
	go expireVotesForever(b)
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
		//TODO(carla): if seltled and missed, update to paid

		if err := votes_db.UpdateStatus(ctx, b.GetDB(), exp.ID, types.VoteStatusCreated,
			types.VoteStatusExpired); err != nil {
			return err
		}
	}
	return nil
}
