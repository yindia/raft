package route

import (
	"context"
	"log"

	v1 "raft/internal/gen/raft/v1"
	"raft/internal/raft"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

//go:generate mockery --output=./mocks --case=underscore --all --with-expecter
type ElectionServiceHandler interface {
	Voting(ctx context.Context, req *connect.Request[v1.VoteRequest]) (*connect.Response[v1.VoteResponse], error)
}

// ElectionServer represents the server handling Redis-like operations.
// It implements the v1.RedisServiceHandler interface.
type ElectionServer struct {
	validator  *protovalidate.Validator
	logger     *log.Logger
	raftServer *raft.RaftServer
}

// NewElectionServer creates and returns a new instance of ElectionServer.
// It initializes the validator and sets up the logger.
func NewElectionServer(raftServer *raft.RaftServer) *ElectionServer {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	server := &ElectionServer{
		validator:  validator,
		logger:     log.New(log.Writer(), "ElectionServer: ", log.LstdFlags|log.Lshortfile),
		raftServer: raftServer,
	}

	server.logger.Println("ElectionServer initialized successfully")
	return server
}

// Join adds a new node to the cluster
func (s *ElectionServer) Voting(ctx context.Context, req *connect.Request[v1.VoteRequest]) (*connect.Response[v1.VoteResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var voteType v1.VoteResponse_VoteType
	if uint64(s.raftServer.CommitIndex()) < req.Msg.LogfileIndex && s.raftServer.Role() != raft.ROLE_LEADER {
		voteType = v1.VoteResponse_VOTE_TYPE_GIVEN
	}

	return connect.NewResponse(&v1.VoteResponse{
		VoteType: voteType,
	}), nil
}
