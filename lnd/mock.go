package lnd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"google.golang.org/grpc"
)

type MockLND struct {
	Client
}

func (m *MockLND) AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error) {
	return &lnrpc.Invoice{PaymentRequest: "test pay req"}, nil
}

func (m *MockLND) AddHoldInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*HoldInvoice, error) {
	var paymentPreimage [32]byte
	rand.Read(paymentPreimage[:])
	hash := sha256.Sum256(paymentPreimage[:])

	return &HoldInvoice{
		Preimage: paymentPreimage[:],
		PayHash:  hex.EncodeToString(hash[:]),
		PayReq:   "pay req",
	}, nil
}

func (m *MockLND) DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error) {
	return new(lnrpc.PayReq), nil
}

type subscribeClient struct {
	grpc.ClientStream
}

func (s *subscribeClient) Recv() (*lnrpc.Invoice, error) {
	return nil, errors.New("force error for test to exit loop")
}

func (m *MockLND) SubscribeInvoice(ctx context.Context, id int64, paymentHash string) (invoicesrpc.Invoices_SubscribeSingleInvoiceClient, error) {
	return &subscribeClient{}, nil
}
