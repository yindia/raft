package route

import (
	"context"
	"log"

	v1 "raft/internal/gen/raft/v1"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type BootstrapServiceHandler interface {
	AddReplica(ctx context.Context, req *connect.Request[v1.AddrInfo]) (*connect.Response[v1.AddrInfoStatus], error)
}

// BootstrapServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type BootstrapServer struct {
	validator *protovalidate.Validator
	logger    *log.Logger
}

// NewBootstrapServer creates and returns a new instance of BootstrapServer.
// It initializes the validator and sets up the logger.
func NewBootstrapServer() *BootstrapServiceHandler {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &BootstrapServer{
		validator: validator,
		logger:    log.New(log.Writer(), "BootstrapServer: ", log.LstdFlags|log.Lshortfile),
	}

	server.logger.Println("BootstrapServer initialized successfully")
	return server
}

// Join adds a new node to the cluster
func (s *BootstrapServer) AddReplica(ctx context.Context, req *connect.Request[v1.AddrInfo]) (*connect.Response[v1.AddrInfoStatus], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&v1.AddrInfoStatus{
		IsAdded: true,
	}), nil
}
