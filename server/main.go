package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-metrics/go-grpc-channelz/server/proto"
	"gitlab.bol.io/kvandenbroek/grpcui/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"time"
)

type greetServer struct {
	proto.UnimplementedGreeterServer
}

func (s greetServer) SayHello(_ context.Context, request *proto.HelloRequest) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{Message: "Hello, " + request.Name}, nil
}

const (
	address    = "localhost:9999"
	webAddress = "localhost:8080"
)

func main() {
	server := grpc.NewServer()
	defer server.GracefulStop()

	reflection.Register(server)
	proto.RegisterGreeterServer(server, greetServer{})
	service.RegisterChannelzServiceToServer(server)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("ERROR: failed to start to start port listener: %s\n", err)
		panic(err)
	}

	go func() {
		log.Println("starting grpc server at " + address)
		if err = server.Serve(listener); err != nil {
			panic(err)
		}
	}()

	err = ServeWebUI(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func ServeWebUI(ctx context.Context) error {
	cc, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("%v, %v", "failed to setup grpc connection", err)
	}

	h, err := standalone.HandlerViaReflection(ctx, cc, address)
	if err != nil {
		return fmt.Errorf("%v, %v", "failed to setup handle via reflection", err)
	}

	webserver := &http.Server{
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           h,
		Addr:              webAddress,
	}

	log.Println("starting web UI at " + webAddress)
	return webserver.ListenAndServe()
}
