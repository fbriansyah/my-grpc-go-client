package hello

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/fbriansyah/my-grpc-go-client/internal/port"
	"github.com/fbriansyah/my-grpc-proto/protogen/go/hello"
	"google.golang.org/grpc"
)

type HelloAdapter struct {
	helloClient port.HelloClientPort
}

func NewHelloAdapter(conn *grpc.ClientConn) (*HelloAdapter, error) {
	client := hello.NewHelloServiceClient(conn)

	return &HelloAdapter{
		helloClient: client,
	}, nil
}

func (a *HelloAdapter) SayHello(ctx context.Context, name string) (*hello.HelloResponse, error) {
	helloRequest := &hello.HelloRequest{Name: "Rian", Age: 19}

	greet, err := a.helloClient.SayHello(ctx, helloRequest)

	if err != nil {
		log.Fatalln(err)
	}

	return greet, nil
}

func (a *HelloAdapter) SayManyHello(ctx context.Context, name string) {
	helloRequest := &hello.HelloRequest{
		Name: "Feb",
	}

	greetStream, err := a.helloClient.SayManyHello(ctx, helloRequest)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		greet, err := greetStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}

		log.Println(greet.Greet)
	}

}

func (a *HelloAdapter) SayHelloToEveryone(ctx context.Context, names []string) {
	greetStream, err := a.helloClient.SayHelloToEveryone(ctx)

	if err != nil {
		log.Fatalln("[SayHelloToEveryone]", err)
	}

	for _, name := range names {
		req := &hello.HelloRequest{
			Name: name,
		}

		greetStream.Send(req)
		time.Sleep(500 * time.Millisecond)
	}

	res, err := greetStream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(res.Greet)
}

func (a *HelloAdapter) SayHelloContinuous(ctx context.Context, names []string) {
	greetStream, err := a.helloClient.SayHelloContinuous(ctx)
	if err != nil {
		log.Fatalln("[SayHelloContinuous]", err)
	}

	greetChan := make(chan struct{})

	go func() {

		for _, name := range names {
			req := &hello.HelloRequest{Name: name}

			greetStream.Send(req)
		}

		greetStream.CloseSend()
	}()

	go func() {
		for {
			greet, err := greetStream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalln("[SayHelloContinuous]", err)
			}

			log.Println(greet.Greet)
		}
		close(greetChan)
	}()

	<-greetChan
}
