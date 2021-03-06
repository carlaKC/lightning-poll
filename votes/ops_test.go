package votes_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/carlaKC/lightning-poll/db"
	"github.com/carlaKC/lightning-poll/lnd"
	"github.com/carlaKC/lightning-poll/votes"
	votes_db "github.com/carlaKC/lightning-poll/votes/internal/db/votes"
	"github.com/carlaKC/lightning-poll/votes/internal/types"
	"github.com/stretchr/testify/assert"
)

var (
	testPollID   = int64(68768)
	testOptionID = int64(45678)
	testSats     = int64(10)
	testExpiry   = int64(100)
	testNote     = "test note"
)

type testBackends struct {
	dbc *sql.DB
	lnd *lnd.MockLND
}

func (b *testBackends) GetDB() *sql.DB {
	return b.dbc
}

func (b *testBackends) GetLND() lnd.Client {
	return b.lnd
}

func setup(t *testing.T) (context.Context, votes.Backends) {
	return context.Background(), &testBackends{dbc: db.ConnectForTesting(t), lnd: &lnd.MockLND{}}
}

func TestCreate(t *testing.T) {
	ctx, b := setup(t)

	_, err := votes.Create(ctx, b, testPollID, testOptionID, testSats, testExpiry, testNote)
	assert.NoError(t, err)
}

func TestGetVotes(t *testing.T) {
	ctx, b := setup(t)

	testOptionID2 := int64(876)

	id1, err := votes.Create(ctx, b, testPollID, testOptionID, testSats, testExpiry, testNote)
	assert.NoError(t, err)
	err = votes_db.UpdateStatus(ctx, b.GetDB(), id1, types.VoteStatusCreated, types.VoteStatusPaid)
	assert.NoError(t, err)

	id2, err := votes.Create(ctx, b, testPollID, testOptionID, testSats, testExpiry, testNote)
	assert.NoError(t, err)
	err = votes_db.UpdateStatus(ctx, b.GetDB(), id2, types.VoteStatusCreated, types.VoteStatusPaid)
	assert.NoError(t, err)

	id3, err := votes.Create(ctx, b, testPollID, testOptionID2, testSats, testExpiry, testNote)
	assert.NoError(t, err)
	err = votes_db.UpdateStatus(ctx, b.GetDB(), id3, types.VoteStatusCreated, types.VoteStatusPaid)
	assert.NoError(t, err)
	_, err = votes.Create(ctx, b, testPollID, testOptionID2, testSats, testExpiry, testNote)
	assert.NoError(t, err)

	v, err := votes.GetResults(ctx, b, testPollID)
	assert.NoError(t, err)
	assert.Equal(t, v[testOptionID], int64(2))
	assert.Equal(t, v[testOptionID2], int64(1))

	_, err = votes.GetResults(ctx, b, testPollID)
	assert.NoError(t, err)

	err = votes_db.UpdateStatus(ctx, b.GetDB(), id1, types.VoteStatusPaid, types.VoteStatusReturned)
	assert.NoError(t, err)
	err = votes_db.UpdateStatus(ctx, b.GetDB(), id2, types.VoteStatusPaid, types.VoteStatusReturned)
	assert.NoError(t, err)

	v, err = votes.GetResults(ctx, b, testPollID)
	assert.NoError(t, err)
	assert.Equal(t, v[testOptionID], int64(2))
	assert.Equal(t, v[testOptionID2], int64(1))
}
