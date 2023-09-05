package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/anthonyhungnguyen/todo-grpc/pkgs/api/proto/v1"
	"google.golang.org/grpc"
)

func RunServer(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)

	if err != nil {
		return err
	}

	// register service
	server := grpc.NewServer()
	v1.RegisterTodoServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("shutting down GRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
