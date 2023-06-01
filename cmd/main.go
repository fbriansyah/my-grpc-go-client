package main

import (
	"context"
	"log"

	"github.com/fbriansyah/my-grpc-go-client/internal/adapter/bank"
	"github.com/fbriansyah/my-grpc-go-client/internal/adapter/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial("localhost:9090", opts...)

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	// helloAdapter, err := hello.NewHelloAdapter(conn)
	bankAdapter, err := bank.NewBankAdapter(conn)

	if err != nil {
		log.Fatalln(err)
	}

	// runSayHello(helloAdapter, "Febrian")
	// runSayManyHellos(helloAdapter, "Rian")

	// runSayHelloContinuous(helloAdapter, []string{"Feb", "Rian", "Nuur", "Rasyiid"})

	runGetCurrentBalance(bankAdapter, "7835697001xxxx")
}

func runSayHello(adapter *hello.HelloAdapter, name string) {
	greet, err := adapter.SayHello(context.Background(), name)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(greet.Greet)
}

func runSayManyHellos(adapter *hello.HelloAdapter, name string) {
	adapter.SayManyHello(context.Background(), name)
}

func runSayHelloToEveryone(adapter *hello.HelloAdapter, names []string) {
	adapter.SayHelloToEveryone(context.Background(), names)
}

func runSayHelloContinuous(adapter *hello.HelloAdapter, names []string) {
	adapter.SayHelloContinuous(context.Background(), names)
}

func runGetCurrentBalance(adapter *bank.BankAdapter, acct string) {
	balance, _ := adapter.GetCurrentBalance(context.Background(), acct)

	log.Println(balance)
}
