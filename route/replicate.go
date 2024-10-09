package route

import (
	"context"
	"log"

	v1 "raft/internal/gen/raft/v1"
	"raft/internal/logfile"
	"raft/internal/raft"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type ReplicateServiceHandler interface {
	CommitOperation(ctx context.Context, req *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error)
	ApplyOperation(ctx context.Context, req *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error)
	ForwardOperation(ctx context.Context, req *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error)
}

// ReplicateServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type ReplicateServer struct {
	validator  *protovalidate.Validator
	logger     *log.Logger
	raftServer *raft.RaftServer
}

// NewReplicateServer creates and returns a new instance of ReplicateServer.
// It initializes the validator and sets up the logger.
func NewReplicateServer(raftServer *raft.RaftServer) *ReplicateServer {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &ReplicateServer{
		validator:  validator,
		logger:     log.New(log.Writer(), "ReplicateServer: ", log.LstdFlags|log.Lshortfile),
		raftServer: raftServer,
	}

	server.logger.Println("ReplicateServer initialized successfully")
	return server
}

// CommitOperation handles committing a transaction.
func (s *ReplicateServer) CommitOperation(ctx context.Context, req *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error) {
	// Validate the request
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logfileFinalIndex, err := s.raftServer.GetLog().CommitOperation(
		int(req.Msg.ExpectedFinalIndex),
		s.raftServer.CommitIndex(),
		&logfile.Transaction{Index: int(req.Msg.Index), Operation: req.Msg.Operation, Term: int(req.Msg.Term)},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.CommitOperationResponse{
		LogfileFinalIndex: int64(logfileFinalIndex), // Example response
	}), nil
}

// ApplyOperation handles applying an operation.
func (s *ReplicateServer) ApplyOperation(ctx context.Context, req *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error) {
	// Validate the request
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	appliedTxn, err := s.raftServer.GetLog().ApplyOperation()
	if err != nil {
		return nil, err
	}
	s.raftServer.CommitIndexInc()
	s.raftServer.ApplyCh(appliedTxn)
	return connect.NewResponse(&v1.ApplyOperationResponse{}), nil
}

// ForwardOperation handles forwarding an operation.
func (s *ReplicateServer) ForwardOperation(ctx context.Context, req *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error) {
	// Validate the request
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&v1.ForwardOperationResponse{}), nil
}
