package votes

import (
	"context"
	"database/sql"
	"github.com/carlaKC/lightning-poll/db"
	"github.com/carlaKC/lightning-poll/votes/internal/types"
	"math/rand"
	"time"
)

var cols = "id, created_at, expires_at, poll_id, option_id, pay_req, payment_hash, preimage, settle_index, settle_amount, status"

type row interface {
	Scan(dest ...interface{}) error
}

func Create(ctx context.Context, dbc *sql.DB, pollID, optionID, expirySeconds int64, payReq, payHash string, preimage []byte) (int64, error) {
	id := rand.Int63()
	expiresAt := time.Now().Add(time.Second * time.Duration(expirySeconds) * -1)

	r, err := dbc.ExecContext(ctx, "insert into votes set id=?, "+
		"created_at=now(),expires_at=?, poll_id=?, option_id=?, pay_req=?,"+
		"payment_hash=?, preimage=?, status=?", id,
		expiresAt, pollID, optionID, payReq, payHash, preimage, types.VoteStatusCreated)
	if err != nil {
		return 0, err
	}

	return id, db.CheckRowsAffected(r, 1)
}

type DBVote struct {
	ID           int64
	CreatedAt    time.Time
	ExpiresAt    time.Time
	PollID       int64
	OptionID     int64
	PayReq       string
	PayHash      string
	Preimage     []byte
	SettleIndex  int64
	SettleAmount int64
	Status       types.VoteStatus
}

func scan(r row) (vote DBVote, err error) {
	var settleIndex, settleAmount sql.NullInt64
	err = r.Scan(&vote.ID, &vote.CreatedAt, &vote.ExpiresAt, &vote.PollID, &vote.OptionID,
		&vote.PayReq, &vote.PayHash, &vote.Preimage, &settleIndex, &settleAmount, &vote.Status)
	if err != nil {
		return vote, err
	}

	if settleIndex.Valid {
		vote.SettleIndex = settleIndex.Int64
	}
	if settleAmount.Valid {
		vote.SettleAmount = settleAmount.Int64
	}

	return vote, nil
}

func list(ctx context.Context, dbc *sql.DB, query string, args ...interface{}) (votes []*DBVote, err error) {
	rows, err := dbc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		vote, err := scan(rows)
		if err != nil {
			return votes, err
		}
		votes = append(votes, &vote)
	}

	return votes, rows.Err()
}

func ListByPollAndStatus(ctx context.Context, dbc *sql.DB, pollID int64, status types.VoteStatus) ([]*DBVote, error) {
	return list(ctx, dbc, "select "+cols+" from votes where poll_id=? and status=?", pollID, status)
}

func UpdateStatus(ctx context.Context, dbc *sql.DB, id int64, fromStatus, toStatus types.VoteStatus) error {
	r, err := dbc.ExecContext(ctx, "update votes set status=? where status=? and "+
		"id=?", toStatus, fromStatus, id)
	if err != nil {
		return err
	}

	return db.CheckRowsAffected(r, 1)
}

func MarkPaid(ctx context.Context, dbc *sql.DB, id, settleAmount int64, settleIndex uint64) error {
	r, err := dbc.ExecContext(ctx, "update votes set status=?, settle_index=?, "+
		"settle_amount=? where id=?", types.VoteStatusPaid, settleIndex, settleAmount, id)
	if err != nil {
		return err
	}

	return db.CheckRowsAffected(r, 1)
}

// ListExpired returns a list of created votes which have expired
func ListExpired(ctx context.Context, dbc *sql.DB) ([]*DBVote, error) {
	return list(ctx, dbc, "select * from votes where expires_at<now() "+
		"and status=?", types.VoteStatusCreated)
}

func Lookup(ctx context.Context, dbc *sql.DB, id int64) (*DBVote, error) {
	row := dbc.QueryRowContext(ctx, "select "+cols+" from votes where id=?", id)
	vote, err := scan(row)
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func LookupByHash(ctx context.Context, dbc *sql.DB, paymentHash string) (*DBVote, error) {
	row := dbc.QueryRowContext(ctx, "select "+cols+" from votes where payment_hash=?", paymentHash)
	vote, err := scan(row)
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func GetLatestSettleIndex(ctx context.Context, dbc *sql.DB) (int64, error) {
	row := dbc.QueryRowContext(ctx, "select coalesce(max(settle_index), 0)"+
		" from votes")

	var index int64
	if err := row.Scan(&index); err != nil {
		return 0, err
	}

	return index, nil
}
