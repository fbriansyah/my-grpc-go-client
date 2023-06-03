package resiliency

import (
	"context"
	"io"
	"log"

	"github.com/fbriansyah/my-grpc-go-client/internal/port"
	resl "github.com/fbriansyah/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
)

type ResiliencyAdapter struct {
	resiliencyClient port.ResiliencyClietnPort
}

func NewResiliencyAdapter(conn *grpc.ClientConn) (*ResiliencyAdapter, error) {
	client := resl.NewResiliencyServiceClient(conn)

	return &ResiliencyAdapter{
		resiliencyClient: client,
	}, nil
}

func (a *ResiliencyAdapter) UnaryResiliency(ctx context.Context, minDelay int32, maxDelay int32, statusCodes []uint32) (*resl.ResiliencyResponse, error) {
	req := &resl.ResiliencyRequest{
		MinDelaySecond: minDelay,
		MaxDelaySecond: maxDelay,
		StatusCodes:    statusCodes,
	}

	res, err := a.resiliencyClient.UnaryResiliency(ctx, req)
	if err != nil {
		log.Println("Error on UnaryResiliency :", err)
		return nil, err
	}

	return res, nil
}

func (a *ResiliencyAdapter) ServerStreamingResiliency(ctx context.Context, minDelay int32, maxDelay int32, statusCodes []uint32) {
	req := &resl.ResiliencyRequest{
		MinDelaySecond: minDelay,
		MaxDelaySecond: maxDelay,
		StatusCodes:    statusCodes,
	}

	reslStream, err := a.resiliencyClient.ServerStreamingResiliency(ctx, req)
	if err != nil {
		log.Fatalln("Error on ServerStreamingResiliency :", err)
	}

	for {
		res, err := reslStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln("Error on ServerStreamingResiliency :", err)
		}

		log.Println(res.DummyString)
	}
}

func (a *ResiliencyAdapter) ClientStreamingResiliency(ctx context.Context, minDelay int32, maxDelay int32, statusCodes []uint32, count int) {
	reslStream, err := a.resiliencyClient.ClientStreamingResiliency(ctx)
	if err != nil {
		log.Fatalln("Error on ClientStreamingResiliency :", err)
	}

	for i := 1; i <= count; i++ {
		req := &resl.ResiliencyRequest{
			MinDelaySecond: minDelay,
			MaxDelaySecond: maxDelay,
			StatusCodes:    statusCodes,
		}
		reslStream.Send(req)
	}

	res, err := reslStream.CloseAndRecv()
	if err != nil {
		log.Fatalln("Error on ClientStreamingResiliency :", err)
	}

	log.Println(res.DummyString)
}

func (a *ResiliencyAdapter) BiDirectionalResiliency(ctx context.Context, minDelay int32, maxDelay int32, statusCodes []uint32, count int) {
	reslStream, err := a.resiliencyClient.BiDirectionalResiliency(ctx)
	if err != nil {
		log.Fatalln("Error on BiDirectionalResiliency :", err)
	}

	reslChan := make(chan struct{})

	go func() {
		for i := 1; i <= count; i++ {
			req := &resl.ResiliencyRequest{
				MinDelaySecond: minDelay,
				MaxDelaySecond: maxDelay,
				StatusCodes:    statusCodes,
			}
			reslStream.Send(req)
		}

		reslStream.CloseSend()
	}()

	go func() {
		for {
			res, err := reslStream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalln("Error on BiDirectionalResiliency :", err)
			}

			log.Println(res.DummyString)
		}
		close(reslChan)
	}()

	<-reslChan
}
