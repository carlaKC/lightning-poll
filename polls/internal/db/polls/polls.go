package polls

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"lightning-poll/db"
	"lightning-poll/polls/internal/types"
	ext_types "lightning-poll/types"
)

var cols = "id, status, created_at,expires_at, question, expiry_seconds, repay_scheme, vote_sats, payout_invoice"

type row interface {
	Scan(dest ...interface{}) error
}

func Create(ctx context.Context, dbc *sql.DB, question, payoutInvoice, email string,
	repayScheme ext_types.RepayScheme, expirySeconds, voteSats int64) (int64, error) {

	id := rand.Int63()
	nullEmail := sql.NullString{String:email, Valid:email!=""}

	r, err := dbc.ExecContext(ctx, "insert into polls set id=?, status=?, "+
		"created_at=now(), expires_at=DATE_ADD(now(), INTERVAL ? SECOND), question=?," +
		" expiry_seconds=?, repay_scheme=?, vote_sats=?, payout_invoice=?, email=?",
		id, types.PollStatusCreated, expirySeconds, question, expirySeconds,
		repayScheme, voteSats, payoutInvoice, nullEmail)
	if err != nil {
		return 0, err
	}

	return id, db.CheckRowsAffected(r, 1)
}

type DBPoll struct {
	ID            int64
	Status        types.PollStatus
	CreatedAt     time.Time
	ExpiresAt     time.Time
	Question      string
	ExpirySeconds int64
	RepayScheme   ext_types.RepayScheme
	VoteSats      int64
	PayoutInvoice string
}

func scan(r row) (poll DBPoll, err error) {
	var invoice sql.NullString

	err = r.Scan(&poll.ID, &poll.Status, &poll.CreatedAt, &poll.ExpiresAt, &poll.Question,
		&poll.ExpirySeconds, &poll.RepayScheme, &poll.VoteSats, &invoice)
	if err != nil {
		return poll, err
	}

	if invoice.Valid {
		poll.PayoutInvoice = invoice.String
	}

	return poll, nil
}

func list(ctx context.Context, dbc *sql.DB, query string, args ...interface{}) (polls []*DBPoll, err error) {
	rows, err := dbc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		poll, err := scan(rows)
		if err != nil {
			return polls, err
		}
		polls = append(polls, &poll)
	}

	return polls, rows.Err()
}

func Lookup(ctx context.Context, dbc *sql.DB, id int64) (*DBPoll, error) {
	row := dbc.QueryRowContext(ctx, "select "+cols+" from polls where id=?", id)
	poll, err := scan(row)
	if err == sql.ErrNoRows {
		return nil, db.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &poll, nil
}

func ListByStatus(ctx context.Context, dbc *sql.DB, status types.PollStatus) ([]*DBPoll, error) {
	return list(ctx, dbc, "select "+cols+" from polls where status=?", status)
}

func UpdateStatus(ctx context.Context, dbc *sql.DB, id int64, fromStatus, toStatus types.PollStatus) error {
	r, err := dbc.ExecContext(ctx, "update polls set status=? where id=? and "+
		"status=?", toStatus, id, fromStatus)
	if err != nil {
		return err
	}

	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return db.ErrUnexpectedRowCount
	}

	return nil
}

// ListExpired returns a list of created votes which have expired
func ListExpired(ctx context.Context, dbc *sql.DB) ([]*DBPoll, error) {
	return list(ctx, dbc, "select "+cols+" from polls where expires_at<now() "+
		"and status=?", types.PollStatusCreated)
}
