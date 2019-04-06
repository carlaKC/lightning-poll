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

func (m *MockLND) AddHoldInvoice(ctx context.Context, amount, expirySeconds int64, note string) (string, error) {
	return "pay req", nil
}

func (m *MockLND) DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error) {
	return new(lnrpc.PayReq), nil
}
