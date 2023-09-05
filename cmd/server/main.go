package main

import (
	"log"
	"os"

	cmd "github.com/anthonyhungnguyen/todo-grpc/pkgs/cmd/server"
)

func main() {
	if err := cmd.RunServer(); err != nil {
		log.Fatalf("failed to server: %v", err)
		os.Exit(1)
	}
}
