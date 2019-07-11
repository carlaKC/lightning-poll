package lnd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	lnd_address   = flag.String("lnd_address", "127.0.0.1:10001", "LND rpc server address")
	lnd_cert      = flag.String("lnd_cert", "/home/lnd/.lnd/tls.cert", "lnd cert location")
	macaroon_path = flag.String("macaroon_path", "/home/lnd/.lnd/admin.macaroon", "LND admin macaroon location")
)

type Client interface {
	AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error)
	AddHoldInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*HoldInvoice, error)
	CancelHoldInvoice(ctx context.Context, hash string) error
	SettleHoldInvoice(ctx context.Context, preimage []byte) error
	LookupInvoice(ctx context.Context, paymentHash string) (*lnrpc.Invoice, error)
	SubscribeInvoice(ctx context.Context, id int64, paymentHash string) (invoicesrpc.Invoices_SubscribeSingleInvoiceClient, error)
	DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error)
	SendPaymentSync(ctx context.Context, payReq string, amount int64) (*lnrpc.SendResponse, error)
}

type client struct {
	rpcConn       *grpc.ClientConn
	rpcClient     lnrpc.LightningClient
	invoiceClient invoicesrpc.InvoicesClient
	macaroon      string
}

type HoldInvoice struct {
	Preimage []byte
	PayHash  string
	PayReq   string
}

// New returns a grpc client which connects to LND's rpc server.
func New() (Client, error) {
	cl := new(client)
	err := cl.connect(*lnd_address, *lnd_cert, "")
	if err != nil {
		return nil, errors.Wrap(err, "cl.connect error")
	}

	dat, err := ioutil.ReadFile(*macaroon_path)
	if err != nil {
		return nil, err
	}
	cl.macaroon = hex.EncodeToString(dat)

	return cl, nil
}

func (cl *client) macaroonCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(
		ctx, "macaroon", cl.macaroon)
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
	cl.invoiceClient = invoicesrpc.NewInvoicesClient(conn)

	return nil
}

func (cl *client) AddInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*lnrpc.Invoice, error) {
	inv := &lnrpc.Invoice{
		Value:  amount,
		Expiry: expirySeconds,
		Memo:   note,
	}

	resp, err := cl.rpcClient.AddInvoice(cl.macaroonCtx(ctx), inv)
	if err != nil {
		return nil, err
	}
	inv.PaymentRequest = resp.PaymentRequest
	inv.AddIndex = resp.AddIndex

	return inv, nil
}

func (cl *client) AddHoldInvoice(ctx context.Context, amount, expirySeconds int64, note string) (*HoldInvoice, error) {
	var paymentPreimage [32]byte
	rand.Read(paymentPreimage[:])
	paymentHash := sha256.Sum256(paymentPreimage[:])

	resp, err := cl.invoiceClient.AddHoldInvoice(
		cl.macaroonCtx(ctx),
		&invoicesrpc.AddHoldInvoiceRequest{
			Value:  amount,
			Expiry: expirySeconds,
			Hash:   paymentHash[:],
			Memo:   note,
		})
	if err != nil {
		return nil, err
	}

	return &HoldInvoice{
		Preimage: paymentPreimage[:],
		PayHash:  hex.EncodeToString(paymentHash[:]),
		PayReq:   resp.PaymentRequest,
	}, nil
}

func (cl *client) CancelHoldInvoice(ctx context.Context, hash string) error {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	_, err = cl.invoiceClient.CancelInvoice(
		cl.macaroonCtx(ctx),
		&invoicesrpc.CancelInvoiceMsg{
			PaymentHash: hashBytes,
		})
	return err
}

func (cl *client) SettleHoldInvoice(ctx context.Context, preimage []byte) error {
	_, err := cl.invoiceClient.SettleInvoice(
		cl.macaroonCtx(ctx),
		&invoicesrpc.SettleInvoiceMsg{Preimage: preimage},
	)
	return err
}

func (cl *client) LookupInvoice(ctx context.Context, paymentHash string) (*lnrpc.Invoice, error) {
	return cl.rpcClient.LookupInvoice(
		cl.macaroonCtx(ctx),
		&lnrpc.PaymentHash{
			RHashStr: paymentHash,
		})
}

func (cl *client) SubscribeInvoice(ctx context.Context, id int64, paymentHash string) (invoicesrpc.Invoices_SubscribeSingleInvoiceClient, error) {
	log.Printf("lnd: SubscribeInvoice connecting for invoice: %v", id)

	hash, err := hex.DecodeString(paymentHash)
	if err != nil {
		return nil, err
	}

	return cl.invoiceClient.SubscribeSingleInvoice(
		cl.macaroonCtx(ctx),
		&invoicesrpc.SubscribeSingleInvoiceRequest{
			RHash: hash,
		})
}

func (cl *client) DecodePaymentRequest(ctx context.Context, request string) (*lnrpc.PayReq, error) {
	return cl.rpcClient.DecodePayReq(
		cl.macaroonCtx(ctx),
		&lnrpc.PayReqString{
			PayReq: request,
		},
	)
}

func (cl *client) SendPaymentSync(ctx context.Context, payReq string, amount int64) (*lnrpc.SendResponse, error) {
	return cl.rpcClient.SendPaymentSync(
		cl.macaroonCtx(ctx),
		&lnrpc.SendRequest{
			PaymentRequest: payReq,
			Amt:            amount,
		})
}
