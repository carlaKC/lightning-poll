package polls

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	poll_db "lightning-poll/polls/internal/db/polls"
	"lightning-poll/polls/internal/types"
	"lightning-poll/votes"
)

func StartLoops(b Backends) {
	go closePollsForever(b)
}

func closePollsForever(b Backends) {
	for {
		if err := closePolls(b); err != nil {
			log.Printf("polls/ops: closePollsForever error: %v", err)
		}
		time.Sleep(time.Minute * 1)
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

	// all votes have been released, update poll's status
	if err := poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusClosed,
		types.PollStatusReleased); err != nil {
		return err
	}

	amount, err := votes.ReleaseVotesForPoll(ctx, b, poll.ID, poll.RepayScheme.GetScheme())
	if err != nil {
		return err
	}

	// the poll creator does not need to be paid out.
	if amount == 0 {
		log.Printf("polls/ops: poll %v has no balance to pay out", poll.ID)
		return poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusReleased,
			types.PollStatusPaidOut)
	}

	// update to paying out to prevent double sending if sync send payment fails
	if err := poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusReleased,
		types.PollStatusPayingOut); err != nil {
		return err
	}

	resp, err := b.GetLND().SendPaymentSync(ctx, poll.PayoutInvoice, amount)
	if err != nil {
		return err
	}
	if resp.PaymentError != "" {
		return fmt.Errorf("polls/ops: closePoll %v error: %v", poll.ID, resp.PaymentError)
	}

	// poll has been paid out, update to final state
	if err := poll_db.UpdateStatus(ctx, b.GetDB(), poll.ID, types.PollStatusPayingOut,
		types.PollStatusPaidOut); err != nil {
		return err
	}

	return nil
}
