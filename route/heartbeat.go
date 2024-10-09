package route

import (
	"context"
	"log" // Remove this import
	raftv1 "raft/internal/gen/raft/v1"
	"raft/internal/raft"
	"strings"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"golang.org/x/exp/slog" // Add this import for slog
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type HeartbeatServiceHandler interface {
	SendHeartbeat(ctx context.Context, req *connect.Request[raftv1.HeartbeatRequest]) (*connect.Response[raftv1.HeartbeatResponse], error)
}

// HeartbeatServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type HeartbeatServer struct {
	validator  *protovalidate.Validator
	logger     *slog.Logger // Change logger type to slog.Logger
	raftServer *raft.RaftServer
}

// NewHeartbeatServer creates and returns a new instance of HeartbeatServer.
// It initializes the validator and sets up the logger.
func NewHeartbeatServer(raftServer *raft.RaftServer) *HeartbeatServer {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &HeartbeatServer{
		validator: validator,
		logger: slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), // Initialize slog logger
		raftServer: raftServer,
	}

	server.logger.Info("HeartbeatServer initialized successfully") // Update logging method
	return server
}

// Join adds a new node to the cluster
func (s *HeartbeatServer) SendHeartbeat(ctx context.Context, req *connect.Request[raftv1.HeartbeatRequest]) (*connect.Response[raftv1.HeartbeatResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		s.logger.Error("Validation failed",
			slog.String("error", err.Error()),
			slog.String("request_addr", req.Msg.Addr)) // Added closing parenthesis
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.raftServer.SetLeader(req.Msg.Addr)
	s.raftServer.JoinMember(req.Msg.Addr)
	for _, n := range req.Msg.Node {
		if n != s.raftServer.Addr() {
			s.raftServer.JoinMember(strings.Split(n, ":")[0])
		}
	}
	if s.raftServer.Heartbeat != nil {
		s.raftServer.Heartbeat.Beat()
		s.logger.Info("Heartbeat sent", slog.String("leader_addr", req.Msg.Addr)) // Log when heartbeat is sent
		return connect.NewResponse(&raftv1.HeartbeatResponse{
			IsAlive: true,
		}), nil
	}

	s.logger.Warn("Heartbeat not sent, no leader found", slog.String("request_addr", req.Msg.Addr)) // Log warning if no leader
	return connect.NewResponse(&raftv1.HeartbeatResponse{
		IsAlive: true,
	}), nil
}
