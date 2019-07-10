package options

import (
	"context"
	"database/sql"
	"math/rand"

	"github.com/carlaKC/lightning-poll/db"
)

var cols = "id, poll_id, value"

type row interface {
	Scan(dest ...interface{}) error
}

func Create(ctx context.Context, dbc *sql.DB, pollID int64, value string) (int64, error) {
	id := rand.Int63()
	r, err := dbc.ExecContext(ctx, "insert into poll_options set id=?, poll_id=?, value=?", id, pollID, value)
	if err != nil {
		return 0, err
	}

	return id, db.CheckRowsAffected(r, 1)
}

type DBOption struct {
	ID     int64
	PollID int64
	Value  string
}

func scan(r row) (option DBOption, err error) {
	err = r.Scan(&option.ID, &option.PollID, &option.Value)
	if err != nil {
		return option, err
	}

	return option, nil
}

func list(ctx context.Context, dbc *sql.DB, query string, args ...interface{}) (options []*DBOption, err error) {
	rows, err := dbc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		opt, err := scan(rows)
		if err != nil {
			return options, err
		}
		options = append(options, &opt)
	}

	return options, rows.Err()
}

func ListByPoll(ctx context.Context, dbc *sql.DB, pollID int64) ([]*DBOption, error) {
	return list(ctx, dbc, "select "+cols+" from poll_options where poll_id=?", pollID)
}
