package route

import (
	"context"
	"log"
	"strings"

	"log/slog"

	v1 "raft/internal/gen/raft/v1"
	"raft/internal/raft"

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
	validator  *protovalidate.Validator
	logger     *slog.Logger // Change logger type to slog.Logger
	raftServer *raft.RaftServer
}

// NewBootstrapServer creates and returns a new instance of BootstrapServer.
// It initializes the validator and sets up the logger.
func NewBootstrapServer(raftServer *raft.RaftServer) *BootstrapServer {
	validator, err := protovalidate.New()
	if err != nil {
		slog.Error("Failed to initialize validator", "error", err) // Use slog for logging
	}

	server := BootstrapServer{
		validator: validator,
		logger: slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), // Initialize slog logger
		raftServer: raftServer,
	}

	server.logger.Info("BootstrapServer initialized successfully") // Use slog for logging
	return &server                                                 // Return a pointer to the server
}

// Join adds a new node to the cluster
func (s *BootstrapServer) AddReplica(ctx context.Context, req *connect.Request[v1.AddrInfo]) (*connect.Response[v1.AddrInfoStatus], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	s.raftServer.JoinMember(strings.Split(req.Msg.Addr, ":")[0])
	var nodes []string
	for _, member := range s.raftServer.Memberlist().Members() {
		nodes = append(nodes, member.Address())
	}
	return connect.NewResponse(&v1.AddrInfoStatus{
		IsAdded: true,
		Nodes:   nodes,
	}), nil
}
