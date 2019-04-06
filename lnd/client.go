package lnd

import (
	"context"
	"flag"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var lnd_address = flag.String("lnd_address", "127.0.0.1:10001", "LND rpc server address")
var lnd_cert = flag.String("lnd_cert", "/home/lnd/.lnd/tls.cert", "lnd cert location")

type Client interface {
	AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error)
	LookupInvoice(ctx context.Context, paymentHash string) (*lnrpc.Invoice, error)
	SubscribeInvoices(ctx context.Context, minSettleIndex int64) (lnrpc.Lightning_SubscribeInvoicesClient, error)
	DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error)
	SendPaymentSync(ctx context.Context, payReq string) (*lnrpc.SendResponse, error)
}

type client struct {
	rpcConn   *grpc.ClientConn
	rpcClient lnrpc.LightningClient
}

// New returns a grpc client which connects to LND's rpc server.
func New() (Client, error) {
	cl := new(client)
	err := cl.connect(*lnd_address, *lnd_cert, "")
	if err != nil {
		return nil, errors.Wrap(err, "cl.connect error")
	}

	return cl, nil
}

func (cl *client) connect(address, cert, serverNameOverride string) error {
	creds, err := credentials.NewClientTLSFromFile(cert, serverNameOverride)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return errors.Wrap(err, "grpc.Dial error")
	}
	cl.rpcConn = conn
	cl.rpcClient = lnrpc.NewLightningClient(conn)

	return nil
}

func (cl *client) AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error) {
	inv := &lnrpc.Invoice{
		Value:  amount,
		Expiry: expirySeconds,
		Memo:   note,
	}

	resp, err := cl.rpcClient.AddInvoice(ctx, inv)
	if err != nil {
		return nil, err
	}
	inv.PaymentRequest = resp.PaymentRequest
	inv.AddIndex = resp.AddIndex

	return inv, nil
}

func (cl *client) LookupInvoice(ctx context.Context, paymentHash string) (*lnrpc.Invoice, error) {
	return cl.rpcClient.LookupInvoice(ctx, &lnrpc.PaymentHash{
		RHashStr: paymentHash,
	})
}

func (cl *client) SubscribeInvoices(ctx context.Context, minSettleIndex int64) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	return cl.rpcClient.SubscribeInvoices(ctx, &lnrpc.InvoiceSubscription{SettleIndex: uint64(minSettleIndex)})
}

func (cl *client) DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error) {
	return cl.rpcClient.DecodePayReq(ctx, &lnrpc.PayReqString{PayReq: request})
}

func (cl *client) SendPaymentSync(ctx context.Context, payReq string) (*lnrpc.SendResponse, error) {
	return cl.rpcClient.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: payReq})
}
