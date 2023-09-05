#/bin/bash
protoc --go_out=./pkgs --go_opt=paths=source_relative \
    --go-grpc_out=./pkgs --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
    api/proto/v1/todo-service.proto