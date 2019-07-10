package polls_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/carlaKC/lightning-poll/db"
	"github.com/carlaKC/lightning-poll/polls/internal/db/polls"
	"github.com/carlaKC/lightning-poll/polls/internal/types"
	ext_types "github.com/carlaKC/lightning-poll/types"
	"github.com/stretchr/testify/assert"
)

var (
	testQuestion = "test question"
	testInvoice  = "lnsb100n1pwfm4pwpp5dn7dgk3h98yqr9lxs79g98tkwl36gxhck6j66n23teftyuje9avqdq8w3jhxaqcqzysxqzfvyhv6jv007k4c05v5xhz2flzjs08j44z02yjex6qp0hrqd4f5sw794jwrhzhfztqkrzprnt755dd6w0zv0cpq5hjgvasr2j4vnhxawygp7v9z5x"
	testRepay    = ext_types.RepaySchemeAll
	testExpiry   = int64(100)
	testVoteSats = int64(10)
	testUser     = int64(123)
)

func setup(t *testing.T) (context.Context, *sql.DB) {
	return context.Background(), db.ConnectForTesting(t)
}

func TestCreate(t *testing.T) {
	ctx, dbc := setup(t)
	_, err := polls.Create(ctx, dbc, testQuestion, testInvoice, "", testRepay, testExpiry, testVoteSats)
	assert.NoError(t, err)
}

func TestLookup(t *testing.T) {
	ctx, dbc := setup(t)
	id, err := polls.Create(ctx, dbc, testQuestion, testInvoice, "", testRepay, testExpiry, testVoteSats)
	assert.NoError(t, err)

	_, err = polls.Lookup(ctx, dbc, id)
	assert.NoError(t, err)
}

func TestListByStatus(t *testing.T) {
	ctx, dbc := setup(t)
	_, err := polls.Create(ctx, dbc, testQuestion, testInvoice, "", testRepay, testExpiry, testVoteSats)
	assert.NoError(t, err)

	pList, err := polls.ListByStatus(ctx, dbc, types.PollStatusCreated)
	assert.NoError(t, err)
	assert.Len(t, pList, 1)

	pList, err = polls.ListByStatus(ctx, dbc, types.PollStatusPaidOut)
	assert.NoError(t, err)
	assert.Len(t, pList, 0)
}

func TestUpdateStatus(t *testing.T) {
	ctx, dbc := setup(t)
	id, err := polls.Create(ctx, dbc, testQuestion, testInvoice, "", testRepay, testExpiry, testVoteSats)
	assert.NoError(t, err)

	err = polls.UpdateStatus(ctx, dbc, id, types.PollStatusCreated, types.PollStatusClosed)
	assert.NoError(t, err)

	err = polls.UpdateStatus(ctx, dbc, id, types.PollStatusClosed, types.PollStatusClosed)
	assert.Equal(t, db.ErrUnexpectedRowCount, err)
}
