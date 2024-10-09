package route

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"raft/internal/raft"

	v1 "raft/internal/gen/raft/v1"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type RedisServiceHandler interface {
	Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	Set(ctx context.Context, req *connect.Request[v1.SetRequest]) (*connect.Response[v1.SetResponse], error)
	Del(ctx context.Context, req *connect.Request[v1.DelRequest]) (*connect.Response[v1.DelResponse], error)

	Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
}

// RedisServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type RedisServer struct {
	validator *protovalidate.Validator
	logger    *log.Logger
	raft      *raft.RaftServer
}

// NewRedisServer creates and returns a new instance of RedisServer.
// It initializes the validator and sets up the logger.
func NewRedisServer(raft *raft.RaftServer) RedisServiceHandler {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &RedisServer{
		validator: validator,
		logger:    log.New(log.Writer(), "RedisServer: ", log.LstdFlags|log.Lshortfile),
		raft:      raft,
	}

	server.logger.Println("RedisServer initialized successfully")
	return server
}

// Get retrieves the value for a given key.
func (s *RedisServer) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	b, _ := json.Marshal("")

	return connect.NewResponse(&v1.GetResponse{Value: b}), nil
}

// Set stores a key-value pair.
func (s *RedisServer) Set(ctx context.Context, req *connect.Request[v1.SetRequest]) (*connect.Response[v1.SetResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	txn, err := s.raft.ConvertToTransaction(fmt.Sprintf("set:%s:%s", req.Msg.Key, req.Msg.Value))
	if err != nil {
		return nil, err
	}

	err = s.raft.PerformTwoPhaseCommit(txn)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.SetResponse{Success: true}), nil
}

// Del deletes one or more keys.
func (s *RedisServer) Del(ctx context.Context, req *connect.Request[v1.DelRequest]) (*connect.Response[v1.DelResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	txn, err := s.raft.ConvertToTransaction(fmt.Sprintf("delete:%s", req.Msg.Keys))
	if err != nil {
		return nil, err
	}

	err = s.raft.PerformTwoPhaseCommit(txn)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.DelResponse{DeletedCount: int32(1)}), nil
}

// Ping checks if the server is responsive.
func (s *RedisServer) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return connect.NewResponse(&v1.PingResponse{Message: "PONG"}), nil
}
