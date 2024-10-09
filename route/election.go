package route

import (
	"context"
	"log"      // Remove this import if not needed
	"log/slog" // Ensure this import is present

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
	logger     *slog.Logger // Change log.Logger to slog.Logger
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
		validator: validator,
		logger: slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), // Initialize slog logger
		raftServer: raftServer,
	}

	server.logger.Info("ElectionServer initialized successfully") // Use slog for logging
	return server
}

// Voting handles the voting process for the election.
func (s *ElectionServer) Voting(ctx context.Context, req *connect.Request[v1.VoteRequest]) (*connect.Response[v1.VoteResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		s.logger.Error("Validation failed", "error", err, "request", req.Msg) // Log validation error with request details
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var voteType v1.VoteResponse_VoteType
	if uint64(s.raftServer.CommitIndex()) <= req.Msg.LogfileIndex && s.raftServer.Role() != raft.ROLE_LEADER {
		voteType = v1.VoteResponse_VOTE_TYPE_GIVEN
		s.logger.Info("Vote granted", "logfileIndex", req.Msg.LogfileIndex, "role", s.raftServer.Role()) // Log vote granted
	} else {
		voteType = v1.VoteResponse_VOTE_TYPE_REFUSED
		s.logger.Info("Vote not granted", "logfileIndex", req.Msg.LogfileIndex, "role", s.raftServer.Role()) // Log vote not granted
	}
	s.logger.Debug("Vote granted", "logfileIndex", req.Msg.LogfileIndex, "role", voteType) // Log vote granted
	return connect.NewResponse(&v1.VoteResponse{
		VoteType: voteType,
	}), nil
}
