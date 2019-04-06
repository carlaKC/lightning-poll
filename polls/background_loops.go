package polls

import (
	"log"

	"golang.org/x/net/context"

	poll_db "lightning-poll/polls/internal/db/polls"
	"lightning-poll/polls/internal/types"
	"lightning-poll/votes"
)

func Start(b Backends) {
	go closePollsForever(b)
}

func closePollsForever(b Backends) {
	for {
		if err := closePolls(b); err != nil {
			log.Printf("polls/ops: closePollsForever error: %v", err)
		}
	}
}

func closePolls(b Backends) error {
	ctx := context.Background()

	polls, err := poll_db.ListExpired(ctx, b.GetDB())
	if err != nil {
		return err
	}

	for _, poll := range polls {
		if err := closePoll(ctx, b, poll); err != nil {
			return err
		}
	}

	return nil
}

// closePoll initiates the poll closing process
// - update the poll to closed, so that it cannot receive any more votes
// - return payments to voters, according to the chosen repayment scheme
// - pay the creator the total remaining
func closePoll(ctx context.Context, b Backends, poll *poll_db.DBPoll) error {
	if err := poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusCreated,
		types.PollStatusClosed); err != nil {
		return err
	}

	results, err := votes.GetResults(ctx, b, poll.ID)
	if err != nil {
		return err
	}

	votes, err := votes.GetVotes(ctx, b, poll.ID)
	if err != nil {
		return err
	}

	shouldRepay := poll.RepayScheme.GetScheme()
	for _, vote := range votes {
		if shouldRepay(results, vote.OptionID) {
			log.Printf("polls/ops: Should be releasing %v", vote)
			continue
		}
		log.Printf("polls/ops: Should be settling %v", vote)
	}

	// all votes have been released, update poll's status
	if err := poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusClosed,
		types.PollStatusReleased); err != nil {
		return err
	}

	//TODO(carla): pay out the remainder to the poll creator

	return nil
}
