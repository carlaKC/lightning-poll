package votes_test

import (
	"context"
	"database/sql"
	"lightning-poll/db"
	"lightning-poll/votes/internal/db/votes"
	"lightning-poll/votes/internal/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testPollID   = int64(54678)
	testOptionID = int64(54678)
	testInvoice  = "lnsb100n1pwfm4pwpp5dn7dgk3h98yqr9lxs79g98tkwl36gxhck6j66n23teftyuje9avqdq8w3jhxaqcqzysxqzfvyhv6jv007k4c05v5xhz2flzjs08j44z02yjex6qp0hrqd4f5sw794jwrhzhfztqkrzprnt755dd6w0zv0cpq5hjgvasr2j4vnhxawygp7v9z5x"
	testPayHash  = "b168b765e28fa49a88991f36e27ffe4cd7dd330baba25752ddad90ef7cb013e6"
)

func setup(t *testing.T) (context.Context, *sql.DB) {
	return context.Background(), db.ConnectForTesting(t)
}

func TestCreate(t *testing.T) {
	ctx, dbc := setup(t)

	_, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)
}

func TestListByPollAndStatus(t *testing.T) {
	ctx, dbc := setup(t)

	_, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)
	_, err = votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)

	vList, err := votes.ListByPollAndStatus(ctx, dbc, testPollID, types.VoteStatusCreated)
	assert.NoError(t, err)
	assert.Len(t, vList, 2)

	vList, err = votes.ListByPollAndStatus(ctx, dbc, testPollID, types.VoteStatusExpired)
	assert.NoError(t, err)
	assert.Len(t, vList, 0)
}

func TestUpdateStatus(t *testing.T) {
	ctx, dbc := setup(t)

	id, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)

	err = votes.UpdateStatus(ctx, dbc, id, types.VoteStatusCreated, types.VoteStatusExpired)
	assert.NoError(t, err)

	err = votes.UpdateStatus(ctx, dbc, id, types.VoteStatusExpired, types.VoteStatusExpired)
	assert.Equal(t, db.ErrUnexpectedRowCount, err)

}

func TestSettle(t *testing.T) {
	ctx, dbc := setup(t)

	id, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)

	err = votes.Settle(ctx, dbc, id, 1, 4)
	assert.NoError(t, err)

	err = votes.Settle(ctx, dbc, id, 1, 4)
	assert.Equal(t, db.ErrUnexpectedRowCount, err)
}

func TestListExpired(t *testing.T) {
	ctx, dbc := setup(t)

	expired, err := votes.ListExpired(ctx, dbc)
	assert.NoError(t, err)
	assert.Len(t, expired, 0)

	id, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)
	r, err := dbc.ExecContext(ctx, "update votes set expires_at=? where id=?", time.Now().Add(time.Hour*-1), id)
	assert.NoError(t, err)
	assert.NoError(t, db.CheckRowsAffected(r, 1))

	expired, err = votes.ListExpired(ctx, dbc)
	assert.NoError(t, err)
	assert.Len(t, expired, 1)
}

func TestGetLatestSettleIndex(t *testing.T) {
	ctx, dbc := setup(t)

	index, err := votes.GetLatestSettleIndex(ctx, dbc)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), index)

	id, err := votes.Create(ctx, dbc, testPollID, testOptionID, 10, testInvoice, testPayHash)
	assert.NoError(t, err)
	err = votes.Settle(ctx, dbc, id, 1, 4)
	assert.NoError(t, err)

	index, err = votes.GetLatestSettleIndex(ctx, dbc)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), index)
}
