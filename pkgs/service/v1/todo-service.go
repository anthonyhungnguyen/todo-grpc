package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	v1 "github.com/anthonyhungnguyen/todo-grpc/pkgs/api/proto/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

type todoServiceServer struct {
	db *sql.DB
}

// constructor
func NewTodoServiceServer(db *sql.DB) v1.TodoServiceServer {
	return &todoServiceServer{db: db}
}

// check-api
func (s *todoServiceServer) checkApi(api string) error {
	if api != apiVersion {
		return status.Errorf(codes.Unimplemented, "unsupported version, current supported version: %s but asked for %s", api, apiVersion)
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *todoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to connect to database > "+err.Error())
	}
	return c, nil
}

// create
func (s *todoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	created_at, err := ptypes.Timestamp(req.Todo.CreatedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "insert_at has invalid format > "+err.Error())
	}

	updated_at, err := ptypes.Timestamp(req.Todo.UpdatedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "update_at has invalid format > "+err.Error())
	}

	res, err := c.ExecContext(ctx, "INSERT INTO todo(`title`, `description`, `created_at`, `updated_at`) VALUES(?, ?, ?, ?)", req.Todo.Title, req.Todo.Description, created_at, updated_at)

	if err != nil {
		return nil, status.Error(codes.Unknown, "unable to insert to database > "+err.Error())
	}

	// Get last ID
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "unable to get latest id > "+err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil

}

// read

func (s *todoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := c.QueryContext(ctx, "SELECT * FROM todo WHERE `id` = ?", req.Id)

	if err != nil {
		return nil, status.Error(codes.Unknown, "unable to get record >"+err.Error())
	}

	defer c.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from Todo"+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with id = %d not found", req.Id))
	}

	var res v1.Todo
	var created_at time.Time
	var updated_at time.Time

	if err := rows.Scan(&res.Id, &res.Title, &res.Description, &created_at, &updated_at); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve field values from Todo row > "+err.Error())
	}

	res.CreatedAt, err = ptypes.TimestampProto(created_at)

	if err != nil {
		return nil, status.Error(codes.Unknown, "created_at field has invalid format > "+err.Error())
	}

	res.UpdatedAt, err = ptypes.TimestampProto(updated_at)
	if err != nil {
		return nil, status.Error(codes.Unknown, "updated_at field has invalid format > "+err.Error())
	}

	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("found multiple Todo rows with id = %d", req.Id))
	}

	return &v1.ReadResponse{
		Api:  apiVersion,
		Todo: &res,
	}, nil
}

// update
func (s *todoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	updated_at, err := ptypes.Timestamp(req.Todo.UpdatedAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "updated_at has invalid format > "+err.Error())
	}

	res, err := c.ExecContext(ctx, "UPDATE todo SET `title` = ?, `description` = ?, `updated_at` = ? WHERE `id` = ?", req.Todo.Title, req.Todo.Description, updated_at, req.Todo.Id)

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update todo"+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value"+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with id = %d not found", req.Todo.Id))
	}

	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

// delete
func (s *todoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res, err := c.ExecContext(ctx, "DELETE FROM todo WHERE `id` = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete todo"+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value"+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with id = %d not found", req.Id))
	}

	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

// readAll
func (s *todoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	rows, err := c.QueryContext(ctx, "SELECT * FROM todo")

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Todo"+err.Error())
	}

	todos := []*v1.Todo{}
	var created_at time.Time
	var updated_at time.Time

	for rows.Next() {
		res := new(v1.Todo)
		if err := rows.Scan(&res.Id, &res.Title, &res.Description, &created_at, &updated_at); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Todo row > "+err.Error())
		}

		res.CreatedAt, err = ptypes.TimestampProto(created_at)
		if err != nil {
			return nil, status.Error(codes.Unknown, "created_at field has invalid format > "+err.Error())
		}
		res.UpdatedAt, err = ptypes.TimestampProto(updated_at)
		if err != nil {
			return nil, status.Error(codes.Unknown, "updated_at field has invalid format > "+err.Error())
		}

		todos = append(todos, res)
	}

	return &v1.ReadAllResponse{
		Api:    apiVersion,
		Todods: todos,
	}, nil
}
