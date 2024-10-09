package route

import (
	"context"
	"log"

	v1 "raft/internal/gen/raft/v1"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type HeartbeatServiceHandler interface {
	SendHeartbeat(ctx context.Context, req *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error)
}

// HeartbeatServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type HeartbeatServer struct {
	validator *protovalidate.Validator
	logger    *log.Logger
}

// NewHeartbeatServer creates and returns a new instance of HeartbeatServer.
// It initializes the validator and sets up the logger.
func NewHeartbeatServer() *HeartbeatServiceHandler {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &HeartbeatServer{
		validator: validator,
		logger:    log.New(log.Writer(), "HeartbeatServer: ", log.LstdFlags|log.Lshortfile),
	}

	server.logger.Println("HeartbeatServer initialized successfully")
	return server
}

// Join adds a new node to the cluster
func (s *HeartbeatServer) SendHeartbeat(ctx context.Context, req *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&v1.HeartbeatResponse{
		Acknowledged: false,
	}), nil
}
