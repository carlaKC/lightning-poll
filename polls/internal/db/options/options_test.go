package options_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/carlaKC/lightning-poll/db"
	"github.com/carlaKC/lightning-poll/polls/internal/db/options"
	"github.com/stretchr/testify/assert"
)

var (
	testPollID = int64(54678)
	testValue  = "yes"
)

func setup(t *testing.T) (context.Context, *sql.DB) {
	return context.Background(), db.ConnectForTesting(t)
}

func TestCreate(t *testing.T) {
	ctx, dbc := setup(t)

	_, err := options.Create(ctx, dbc, testPollID, testValue)
	assert.NoError(t, err)
}

func TestListByPoll(t *testing.T) {
	ctx, dbc := setup(t)

	testPollID2 := int64(3454)

	_, err := options.Create(ctx, dbc, testPollID, testValue)
	assert.NoError(t, err)
	_, err = options.Create(ctx, dbc, testPollID, testValue)
	assert.NoError(t, err)
	_, err = options.Create(ctx, dbc, testPollID2, testValue)
	assert.NoError(t, err)

	opts, err := options.ListByPoll(ctx, dbc, testPollID)
	assert.NoError(t, err)
	assert.Len(t, opts, 2)

	opts, err = options.ListByPoll(ctx, dbc, testPollID2)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
}
