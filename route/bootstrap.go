package route

import (
	"context"
	"log"
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
	logger     *log.Logger
	raftServer *raft.RaftServer
}

// NewBootstrapServer creates and returns a new instance of BootstrapServer.
// It initializes the validator and sets up the logger.
func NewBootstrapServer(raftServer *raft.RaftServer) *BootstrapServer {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := BootstrapServer{
		validator:  validator,
		logger:     log.New(log.Writer(), "BootstrapServer: ", log.LstdFlags|log.Lshortfile),
		raftServer: raftServer,
	}

	server.logger.Println("BootstrapServer initialized successfully")
	return &server // Return a pointer to the server
}

// Join adds a new node to the cluster
func (s *BootstrapServer) AddReplica(ctx context.Context, req *connect.Request[v1.AddrInfo]) (*connect.Response[v1.AddrInfoStatus], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if ok := s.raftServer.ReplicaBootstrapConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		log.Printf("raft server [%s] present in replicaConnMap", req.Msg.Addr)
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaBootstrapConnMapLock.Lock()
	s.raftServer.ReplicaBootstrapConnMap[req.Msg.Addr] = raftv1connect.NewBootstrapServiceClient(http.DefaultClient, req.Msg.Addr)
	s.raftServer.ReplicaBootstrapConnMapLock.Unlock()

	if ok := s.raftServer.ReplicaElectionConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		log.Printf("raft server [%s] present in replicaConnMap", req.Msg.Addr)
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaElectionConnMapLock.Lock()
	s.raftServer.ReplicaElectionConnMap[req.Msg.Addr] = raftv1connect.NewElectionServiceClient(http.DefaultClient, req.Msg.Addr)
	s.raftServer.ReplicaElectionConnMapLock.Unlock()

	if ok := s.raftServer.ReplicaHeartbeatConnMap[req.Msg.Addr]; ok != nil {
		// raft server already present
		log.Printf("raft server [%s] present in replicaConnMap", req.Msg.Addr)
		return connect.NewResponse(&v1.AddrInfoStatus{
			IsAdded: false,
		}), nil
	}

	s.raftServer.ReplicaHeartbeatConnMapLock.Lock()
	s.raftServer.ReplicaHeartbeatConnMap[req.Msg.Addr] = raftv1connect.NewHeartbeatServiceClient(http.DefaultClient, req.Msg.Addr)
	s.raftServer.ReplicaHeartbeatConnMapLock.Unlock()

	return connect.NewResponse(&v1.AddrInfoStatus{
		IsAdded: true,
	}), nil
}
