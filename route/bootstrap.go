package route

import (
	"context"
	"fmt"
	"log"

	"log/slog"
	"net/http"

	v1 "raft/internal/gen/raft/v1"
	"raft/internal/gen/raft/v1/raftv1connect"
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
	nodes := s.raftServer.BootstrapNodes
	s.raftServer.BootstrapNodes = append(s.raftServer.BootstrapNodes, req.Msg.Addr)
	if ok := s.raftServer.ReplicaBootstrapConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		s.logger.Warn("raft server already present in replicaConnMap", "address", req.Msg.Addr) // Use slog for logging
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaBootstrapConnMapLock.Lock()
	s.raftServer.ReplicaBootstrapConnMap[req.Msg.Addr] = raftv1connect.NewBootstrapServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", req.Msg.Addr))
	s.raftServer.ReplicaBootstrapConnMapLock.Unlock()

	if ok := s.raftServer.ReplicaElectionConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		s.logger.Warn("raft server already present in replicaConnMap", "address", req.Msg.Addr) // Use slog for logging
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaElectionConnMapLock.Lock()
	s.raftServer.ReplicaElectionConnMap[req.Msg.Addr] = raftv1connect.NewElectionServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", req.Msg.Addr))
	s.raftServer.ReplicaElectionConnMapLock.Unlock()

	if ok := s.raftServer.ReplicaHeartbeatConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		s.logger.Warn("raft server already present in replicaConnMap", "address", req.Msg.Addr) // Use slog for logging
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaHeartbeatConnMapLock.Lock()
	s.raftServer.ReplicaHeartbeatConnMap[req.Msg.Addr] = raftv1connect.NewHeartbeatServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", req.Msg.Addr))
	s.raftServer.ReplicaHeartbeatConnMapLock.Unlock()

	if ok := s.raftServer.ReplicaReplicateConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		s.logger.Warn("raft server already present in replicaConnMap", "address", req.Msg.Addr) // Use slog for logging
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaReplicateConnMapLock.Lock()
	s.raftServer.ReplicaReplicateConnMap[req.Msg.Addr] = raftv1connect.NewReplicateOperationServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", req.Msg.Addr))
	s.raftServer.ReplicaReplicateConnMapLock.Unlock()

	return connect.NewResponse(&v1.AddrInfoStatus{
		IsAdded: true,
		Nodes:   nodes,
	}), nil
}
