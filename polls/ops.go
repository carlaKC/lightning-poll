package polls

import (
	"context"
	"database/sql"
	"lightning-poll/lnd"
	"log"

	"github.com/pkg/errors"

	options_db "lightning-poll/polls/internal/db/options"
	poll_db "lightning-poll/polls/internal/db/polls"
	"lightning-poll/polls/internal/types"
	ext_types "lightning-poll/types"
)

var expiryBufferSeconds int64 = 60 * 60 * 12 // 12 hours in seconds

type Backends interface {
	GetDB() *sql.DB
	GetLND() lnd.Client
}

var (
	ErrNonZeroInvoice = errors.New("Payout invoice is non-zero")
	ErrPayoutExpiry   = errors.New("Payout invoice expires too soon")
)

func CreatePoll(ctx context.Context, b Backends, question, payReq string, repayScheme ext_types.RepayScheme, options []string, expirySeconds, voteSats, userID int64) (int64, error) {
	if err := validatePayout(ctx, b, payReq, expirySeconds); err != nil {
		return 0, err
	}

	id, err := poll_db.Create(ctx, b.GetDB(), question, payReq, repayScheme, expirySeconds, voteSats, userID)
	if err != nil {
		return 0, err
	}

	log.Printf("polls/ops: Created poll: %v", id)

	for _, o := range options {
		if o == "" {
			continue
		}
		optID, err := options_db.Create(ctx, b.GetDB(), id, o)
		if err != nil {
			return 0, err
		}
		log.Printf("polls/ops: Created option: %v for poll: %v", optID, id)
	}

	return id, nil
}

// validatePayout ensures that the payout invoice provided by the poll creator
// has a 0 amount, so we can specify any payment amount and that it has a sufficient
// expiry buffer so that it does not expire before we can pay them out.
func validatePayout(ctx context.Context, b Backends, payReq string, expirySeconds int64) error {
	req, err := b.GetLND().DecodePaymentRequest(ctx, payReq)
	if err != nil {
		return err
	}

	if req.Expiry < (expirySeconds + expiryBufferSeconds) {
		return ErrPayoutExpiry
	}

	if req.NumSatoshis != 0 {
		return ErrNonZeroInvoice
	}

	return nil
}

func LookupPoll(ctx context.Context, b Backends, id int64) (*Poll, error) {
	dbPoll, err := poll_db.Lookup(ctx, b.GetDB(), id)
	if err != nil {
		return nil, err
	}

	poll := &Poll{
		ID:       dbPoll.ID,
		Question: dbPoll.Question,
		Cost:     dbPoll.VoteSats,
		ClosesAt: dbPoll.ExpiresAt,
	}

	options, err := options_db.ListByPoll(ctx, b.GetDB(), dbPoll.ID)
	if err != nil {
		return nil, err
	}

	for _, o := range options {
		poll.Options = append(poll.Options, &Option{ID: o.ID, Value: o.Value})
	}

	return poll, nil
}

func ListActivePolls(ctx context.Context, b Backends) ([]*Poll, error) {
	polls, err := poll_db.ListByStatus(ctx, b.GetDB(), types.PollStatusCreated)
	if err != nil {
		return nil, err
	}

	return getList(ctx, b, polls)
}

func ListInactivePolls(ctx context.Context, b Backends) ([]*Poll, error) {
	closed, err := poll_db.ListByStatus(ctx, b.GetDB(), types.PollStatusClosed)
	if err != nil {
		return nil, err
	}
	paidOut, err := poll_db.ListByStatus(ctx, b.GetDB(), types.PollStatusPaidOut)
	if err != nil {
		return nil, err
	}

	return getList(ctx, b, append(paidOut, closed...))
}

func getList(ctx context.Context, b Backends, polls []*poll_db.DBPoll) ([]*Poll, error) {
	var pollList []*Poll
	for _, poll := range polls {
		p, err := LookupPoll(ctx, b, poll.ID)
		if err != nil {
			return nil, err
		}
		pollList = append(pollList, p)
	}

	return pollList, nil
}

func FoceClosePoll(ctx context.Context, b Backends, id int64) error {
	log.Printf("polls.ops: FoceClosePoll closing poll %v", id)

	dbPoll, err := poll_db.Lookup(ctx, b.GetDB(), id)
	if err != nil {
		return err
	}

	return ClosePoll(ctx, b, dbPoll)
}
