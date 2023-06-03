package main

import (
	"context"
	"log"
	"time"

	"github.com/fbriansyah/my-grpc-go-client/internal/adapter/bank"
	"github.com/fbriansyah/my-grpc-go-client/internal/adapter/hello"
	"github.com/fbriansyah/my-grpc-go-client/internal/adapter/resiliency"
	dresl "github.com/fbriansyah/my-grpc-go-client/internal/application/domain/resiliency"
	"github.com/fbriansyah/my-grpc-go-client/internal/interceptor"
	resl_proto "github.com/fbriansyah/my-grpc-proto/protogen/go/resiliency"

	// grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

var cbreaker *gobreaker.CircuitBreaker

func init() {
	mybreaker := gobreaker.Settings{
		Name: "course-circuit-breaker",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)

			log.Printf("Circuit breaker failure is %v, requests is %v, means failure ratio : %v\n",
				counts.TotalFailures, counts.Requests, failureRatio)

			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		Timeout:     4 * time.Second,
		MaxRequests: 3,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit breaker %v changed state, from %v to %v\n\n", name, from, to)
		},
	}

	cbreaker = gobreaker.NewCircuitBreaker(mybreaker)
}

func main() {
	var opts []grpc.DialOption

	// opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// opts = append(opts,
	// 	grpc.WithUnaryInterceptor(
	// 		grpc_retry.UnaryClientInterceptor(
	// 			grpc_retry.WithCodes(codes.Unknown, codes.Internal),
	// 			grpc_retry.WithMax(4),
	// 			grpc_retry.WithBackoff(grpc_retry.BackoffExponential(2*time.Second)),
	// 		),
	// 	),
	// )

	// opts = append(opts,
	// 	grpc.WithStreamInterceptor(
	// 		grpc_retry.StreamClientInterceptor(
	// 			grpc_retry.WithCodes(codes.Unknown, codes.Internal),
	// 			grpc_retry.WithMax(4),
	// 			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(3*time.Second)),
	// 		),
	// 	),
	// )
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(
			interceptor.LogUnaryClientInterceptor(),
			interceptor.BasicUnaryClientInterceptor(),
			interceptor.TimeoutUnaryClientInterceptor(time.Second*5),
		),
	)

	opts = append(opts,
		grpc.WithChainStreamInterceptor(
			interceptor.LogStreamClientInterceptor(),
			interceptor.BasicClientStreamInterceptor(),
			interceptor.TimeoutStreamClientInterceptor(time.Second*20),
		),
	)

	conn, err := grpc.Dial("localhost:9090", opts...)

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	// helloAdapter, err := hello.NewHelloAdapter(conn)
	// bankAdapter, err := bank.NewBankAdapter(conn)
	resiliencyAdapter, err := resiliency.NewResiliencyAdapter(conn)

	if err != nil {
		log.Fatalln(err)
	}

	// runSayHello(helloAdapter, "Febrian")
	// runSayManyHellos(helloAdapter, "Rian")

	// runSayHelloContinuous(helloAdapter, []string{"Feb", "Rian", "Nuur", "Rasyiid"})

	// runGetCurrentBalance(bankAdapter, "7835697001xxxx")
	runUnaryResiliencyWithTimeout(resiliencyAdapter, 3, 4, []uint32{dresl.OK}, time.Second*2)
	// for i := 0; i < 300; i++ {
	// 	runUnaryResiliencyWithCircuitBreaker(resiliencyAdapter, 0, 0, []uint32{dresl.UNKNOWN, dresl.OK})
	// 	time.Sleep(time.Second)
	// }
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

func runUnaryResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelay int32,
	maxDelay int32, statusCodes []uint32, timeout time.Duration) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := adapter.UnaryResiliency(ctx, minDelay, maxDelay, statusCodes)
	if err != nil {
		log.Fatalln("Failed to call UnaryResiliency :", err)
	}
	log.Println(res.DummyString)
}

func runServerStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelay int32,
	maxDelay int32, statusCodes []uint32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	adapter.ServerStreamingResiliency(ctx, minDelay, maxDelay, statusCodes)
}

func runClientStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelay int32,
	maxDelay int32, statusCodes []uint32, count int, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	adapter.ClientStreamingResiliency(ctx, minDelay, maxDelay, statusCodes, count)
}

func runBiDirectionalResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32,
	count int, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	adapter.BiDirectionalResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes, count)
}

func runUnaryResiliency(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32,
	maxDelaySecond int32, statusCodes []uint32) {
	res, err := adapter.UnaryResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)

	if err != nil {
		log.Fatalln("Failed to call UnaryResiliency :", err)
	}

	log.Println(res.DummyString)
}

func runServerStreamingResiliency(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) {
	adapter.ServerStreamingResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
}

func runClientStreamingResiliency(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32,
	count int) {
	adapter.ClientStreamingResiliency(context.Background(), minDelaySecond,
		maxDelaySecond, statusCodes, count)
}

func runBiDirectionalResiliency(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32,
	count int) {
	adapter.BiDirectionalResiliency(context.Background(), minDelaySecond,
		maxDelaySecond, statusCodes, count)
}

func runUnaryResiliencyWithCircuitBreaker(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) {
	cbreakerRes, cbreakerErr := cbreaker.Execute(
		func() (interface{}, error) {
			return adapter.UnaryResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
		},
	)

	if cbreakerErr != nil {
		log.Println("Failed to call UnaryResiliency :", cbreakerErr)
	} else {
		log.Println(cbreakerRes.(*resl_proto.ResiliencyResponse).DummyString)
	}
}

func runUnaryResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32,
	maxDelaySecond int32, statusCodes []uint32) {
	res, err := adapter.UnaryResiliencyWithMetadata(context.Background(),
		minDelaySecond, maxDelaySecond, statusCodes)

	if err != nil {
		log.Fatalln("Failed to call UnaryResiliencyWithMetadata :", err)
	}

	log.Println(res.DummyString)
}

func runServerStreamingResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) {
	adapter.ServerStreamingResiliencyWithMetadata(context.Background(), minDelaySecond,
		maxDelaySecond, statusCodes)
}

func runClientStreamingResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32,
	count int) {
	adapter.ClientStreamingResiliencyWithMetadata(context.Background(), minDelaySecond,
		maxDelaySecond, statusCodes, count)
}

func runBiDirectionalResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter,
	minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32,
	count int) {
	adapter.BiDirectionalResiliencyWithMetadata(context.Background(), minDelaySecond,
		maxDelaySecond, statusCodes, count)
}
