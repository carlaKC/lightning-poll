package lnd

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
)

type MockLND struct {
	Client
}

func (m *MockLND) AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error) {
	return &lnrpc.Invoice{PaymentRequest: "test pay req"}, nil
}
