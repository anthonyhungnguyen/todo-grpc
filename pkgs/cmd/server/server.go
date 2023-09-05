package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	"github.com/anthonyhungnguyen/todo-grpc/pkgs/protocol/grpc"
	v1 "github.com/anthonyhungnguyen/todo-grpc/pkgs/service/v1"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	GRPCPort string
	DBHost   string
	DBUser   string
	DBPass   string
	DBSChema string
}

func RunServer() error {
	ctx := context.Background()

	// get configuration
	var config Config
	flag.StringVar(&config.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&config.DBHost, "db-host", "", "Database host")
	flag.StringVar(&config.DBUser, "db-user", "", "Database user")
	flag.StringVar(&config.DBPass, "db-pass", "", "Database password")
	flag.StringVar(&config.DBSChema, "db-schema", "", "Database schema")
	flag.Parse()

	if len(config.GRPCPort) == 0 {
		return errors.New("invalid TCP port for gRPC server")
	}

	// mysql parse time
	param := "parseTime=true"

	conn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", config.DBUser, config.DBPass, config.DBHost, config.DBSChema, param)

	db, err := sql.Open("mysql", conn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	defer db.Close()

	v1API := v1.NewTodoServiceServer(db)

	return grpc.RunServer(ctx, v1API, config.GRPCPort)
}
