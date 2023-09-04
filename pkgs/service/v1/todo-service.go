package v1

import (
	"context"
	"database/sql"

	v1 "github.com/anthonyhungnguyen276/todo-grpc/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

type TodoService struct {
	db *sql.DB
}

// constructor
func CreateNewTodoService(db *sql.DB) *TodoService {
	return &TodoService{db}
}

// check-api
func (s *TodoService) checkApi(api string) error {
	if api != apiVersion {
		return status.Errorf(codes.Unimplemented, "unsupported version, current supported version: %s but asked for %s", api, apiVersion)
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *TodoService) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to connect to database > "+err.Error())
	}
	return c, nil
}

// create
func (s *TodoService) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	err := s.checkApi(req.Api)
	if err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	created_at, err := ptypes.Timestamp(req.ToDo.InsertAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "insert_at has invalid format > "+err.Error())
	}

	updated_at, err := ptypes.Timestamp(req.ToDo.UpdateAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "update_at has invalid format > "+err.Error())
	}

	res, err := c.ExecContext(ctx, "INSERT INTO todo(`title`, `description`, `created_at`, `updated_at`) VALUES(?, ?, ?, ?)", req.ToDo.Title, req.ToDo.Description, created_at, updated_at)

	if err != nil {
		return nil, status.Error(codes.Unknown, "unable to insert to database > "+err.Error())
	}

	// Get last ID
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "unable to get latest id > "+err.Error())
	}

	return &v1.CreateResponse{
		api: apiVersion,
		id:  id,
	}

}

// read
// update
// delete
// readAll
