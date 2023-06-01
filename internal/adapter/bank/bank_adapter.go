package bank

import (
	"context"
	"io"
	"log"

	dbank "github.com/fbriansyah/my-grpc-go-client/internal/application/domain/bank"
	"github.com/fbriansyah/my-grpc-go-client/internal/port"
	"github.com/fbriansyah/my-grpc-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BankAdapter struct {
	bankClient port.BankClientPort
}

func NewBankAdapter(conn *grpc.ClientConn) (*BankAdapter, error) {
	client := bank.NewBankServiceClient(conn)

	return &BankAdapter{
		bankClient: client,
	}, nil
}

func (a *BankAdapter) GetCurrentBalance(ctx context.Context, acct string) (*bank.CurrentBalanceResponse, error) {
	req := &bank.CurrentBalanceRequest{
		AccountNumber: acct,
	}

	resp, err := a.bankClient.GetCurrentBalance(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalf("Code: %v | %v\n", st.Code().String(), st.Message())

		return resp, err
	}

	return resp, nil
}

func (a *BankAdapter) FetchExchangeRates(ctx context.Context, fromCur, toCur string) {

	req := &bank.ExchangeRateRequest{
		FromCurrency: fromCur,
		ToCurrency:   toCur,
	}

	// open stream to server
	exchangeStream, err := a.bankClient.FetchExchangeRates(ctx, req)

	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalf("[Fatal] error FetchExchangeRates: %v | %v\n", st.Code().String(), st.Message())
	}

	for {
		rate, err := exchangeStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			st, _ := status.FromError(err)

			if st.Code() == codes.InvalidArgument {
				log.Fatalln("[Fatal] error FetchExchangeRates :", st.Message())
			}
		}

		log.Printf(
			"Rate at %v from %v to %v is %v \n",
			rate.Timestamp,
			rate.FromCurrency,
			rate.ToCurrency,
			rate.Rate,
		)
	}

}

func (a *BankAdapter) SummarizeTransactions(ctx context.Context, acct string, tx []*dbank.Transaction) {
	txStreamm, err := a.bankClient.SummarizeTransactions(ctx)

	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalf("[Fatal] error SummarizeTransactions: %v", st.Err())
	}

	for _, t := range tx {
		ttype := bank.TransactionType_TRANSACTION_TYPE_UNSPECIFIED

		if t.TransactionType == dbank.TransactionTypeIn {
			ttype = bank.TransactionType_TRANSACTION_TYPE_IN
		} else if t.TransactionType == dbank.TransactionTypeOut {
			ttype = bank.TransactionType_TRANSACTION_TYPE_OUT
		}

		req := &bank.Transaction{
			AccountNumber: acct,
			Type:          ttype,
			Amount:        t.Amount,
			Notes:         t.Notes,
		}
		txStreamm.Send(req)
	}

	summary, err := txStreamm.CloseAndRecv()
	if err != nil {
		st, _ := status.FromError(err)

		log.Fatalln("[Fatal] error SummarizeTransactions :", st)

	}

	log.Println("Summary:", summary)
}

func (a *BankAdapter) TransferMultiple(ctx context.Context, trf []dbank.TransferTransaction) {
	trfStream, err := a.bankClient.TransferMultiple(ctx)

	if err != nil {
		log.Fatalln("[FATAL] Error on TransferMultiple :", err)
	}

	trfChan := make(chan struct{})

	go func() {
		for _, tt := range trf {
			req := &bank.TransferRequest{
				FromAccountNumber: tt.FromAccountNumber,
				ToAccountNumber:   tt.ToAccountNumber,
				Currency:          tt.Currency,
				Amount:            tt.Amount,
			}

			trfStream.Send(req)
		}

		trfStream.CloseSend()
	}()

	go func() {
		for {
			res, err := trfStream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				handleTransferErrorGrpc(err)
				break
			} else {
				log.Printf("Transfer status %v on %v\n", res.Status, res.Timestamp)
			}
		}

		close(trfChan)
	}()

	<-trfChan
}

func handleTransferErrorGrpc(err error) {
	st := status.Convert(err)

	log.Printf("Error %v on TransferMultiple : %v", st.Code(), st.Message())

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.PreconditionFailure:
			for _, violation := range t.GetViolations() {
				log.Println("[VIOLATION]", violation)
			}
		case *errdetails.ErrorInfo:
			log.Printf("Error on : %v, with reason %v\n", t.Domain, t.Reason)
			for k, v := range t.GetMetadata() {
				log.Printf("  %v : %v\n", k, v)
			}
		}
	}
}
