package raft

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	raftv1 "raft/internal/gen/raft/v1"
	"raft/internal/gen/raft/v1/raftv1connect"
	"raft/internal/heartbeat"
	"raft/internal/logfile"

	"connectrpc.com/connect"
	"golang.org/x/exp/slog" // Import the slog package
)

const (
	SNAPSHOTS_DIR = "snapshots"
)

// ROLE represents the role of the Raft server in the cluster.
type ROLE int

const (
	ROLE_LEADER    = 1 // Leader role
	ROLE_FOLLOWER  = 2 // Follower role
	ROLE_CANDIDATE = 3 // Candidate role
)

const (
	HEARTBEAT_PERIOD  = time.Second * 10 // Duration between heartbeats
	HEARTBEAT_TIMEOUT = time.Second * 10 // Timeout for heartbeat responses
)

// Initialize the logger
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

// RaftServerOpts holds the options for creating a new Raft server.
type RaftServerOpts struct {
	Address        string   // Address of the server
	Role           ROLE     // Role of the server
	BootstrapNodes []string // List of bootstrap nodes for the cluster
	LeaderAddr     string   // Address of the current leader
}

// RaftServer represents a Raft server instance.
type RaftServer struct {
	role           ROLE                      // Current role of the server
	localAddr      string                    // Address of the server
	BootstrapNodes []string                  // List of bootstrap nodes
	logfile        logfile.Log               // Log file for the server
	applyCh        chan *logfile.Transaction // Channel for applying transactions
	Transport      *http.Server              // HTTP server for transport
	leaderAddr     string                    // Address of the current leader
	Heartbeat      *heartbeat.Heartbeat      // Heartbeat for the server
	currentTerm    int                       // Current term of the server
	commitIndex    int                       // Index of the last committed entry

	// ReplicaConnMap maintains connections to other replicas to reduce latency.
	ReplicaElectionConnMap     ReplicaConnMap[string, raftv1connect.ElectionServiceClient]
	ReplicaElectionConnMapLock sync.RWMutex

	ReplicaHeartbeatConnMap     ReplicaConnMap[string, raftv1connect.HeartbeatServiceClient]
	ReplicaHeartbeatConnMapLock sync.RWMutex

	ReplicaBootstrapConnMap     ReplicaConnMap[string, raftv1connect.BootstrapServiceClient]
	ReplicaBootstrapConnMapLock sync.RWMutex

	ReplicaReplicateConnMap     ReplicaConnMap[string, raftv1connect.ReplicateOperationServiceClient]
	ReplicaReplicateConnMapLock sync.RWMutex
}

type ReplicaConnMap[K comparable, V any] map[K]V

// NewRaftServer creates a new Raft server instance with the provided options.
func NewRaftServer(opts RaftServerOpts) (*RaftServer, map[string]string) {
	raftServer := &RaftServer{
		role:           opts.Role,
		BootstrapNodes: opts.BootstrapNodes,
		Heartbeat:      heartbeat.NewHeartbeat(HEARTBEAT_PERIOD, func() {}),
		logfile:        logfile.NewLogfile(),
		applyCh:        make(chan *logfile.Transaction),
		localAddr:      opts.Address,
		commitIndex:    0,

		ReplicaElectionConnMap:  make(ReplicaConnMap[string, raftv1connect.ElectionServiceClient]),
		ReplicaHeartbeatConnMap: make(ReplicaConnMap[string, raftv1connect.HeartbeatServiceClient]),
		ReplicaBootstrapConnMap: make(ReplicaConnMap[string, raftv1connect.BootstrapServiceClient]),
		ReplicaReplicateConnMap: make(ReplicaConnMap[string, raftv1connect.ReplicateOperationServiceClient]),
	}
	filePath := fmt.Sprintf("%s/%s.%s", SNAPSHOTS_DIR, opts.Address, FILE_EXTENSION)
	_, err := os.Stat(filePath)
	if err != nil {
		logger.Info(fmt.Sprintf("[%s] Snapshot not found. Creating a new RaftServer instance.", opts.Address))
		return raftServer, make(map[string]string)
	}
	logger.Info(fmt.Sprintf("[%s] Restoring RaftServer from snapshot.", opts.Address))

	// If a snapshot exists, restore additional configurations to the server.
	snapshotContent, err := readFile(SNAPSHOTS_DIR, opts.Address)
	if err != nil {
		logger.Error("Error while reading snapshot", "error", err)
	}
	var kvMap map[string]string
	raftServer.commitIndex, kvMap = destructureSnapshot(snapshotContent)
	return raftServer, kvMap
}

// Start initializes the Raft server and begins the heartbeat process.
func (s *RaftServer) Start() error {
	time.Sleep(time.Second * 3) // Wait for the server to start

	// Send requests to bootstrapped servers to add this server to their `replicaConnMap`.
	go s.bootstrapNetwork()

	s.startHeartbeatTimeoutProcess()

	logger.Info("Raft server started successfully.")
	return nil
}

func (s *RaftServer) CommitIndex() int {
	return s.commitIndex
}

func (s *RaftServer) CommitIndexInc() {
	s.commitIndex++
}

func (s *RaftServer) ApplyCh(appliedTxn *logfile.Transaction) {
	s.applyCh <- appliedTxn
}

func (s *RaftServer) SetLeader(addr string) {
	s.leaderAddr = addr
}

func (s *RaftServer) GetLog() logfile.Log {
	return s.logfile
}

func (s *RaftServer) Role() int {
	return int(s.role)
}

func (s *RaftServer) Addr() string {
	return s.localAddr
}

// bootstrapNetwork sends requests to other replicas to add this server to their replicaConnMap.
func (s *RaftServer) bootstrapNetwork() {
	ticker := time.NewTicker(30 * time.Second) // Run every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.connectToBootstrapNodes()
		}
	}
}

func (s *RaftServer) connectToBootstrapNodes() {
	wg := &sync.WaitGroup{}
	for _, addr := range s.BootstrapNodes {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if len(addr) == 0 {
				logger.Warn("Skipping empty bootstrap address.")
				return
			}
			logger.Info(fmt.Sprintf("Attempting to connect with bootstrap node [%s].", addr))

			s.ReplicaElectionConnMap = make(ReplicaConnMap[string, raftv1connect.ElectionServiceClient])
			s.ReplicaHeartbeatConnMap = make(ReplicaConnMap[string, raftv1connect.HeartbeatServiceClient])
			s.ReplicaBootstrapConnMap = make(ReplicaConnMap[string, raftv1connect.BootstrapServiceClient])
			s.ReplicaReplicateConnMap = make(ReplicaConnMap[string, raftv1connect.ReplicateOperationServiceClient])

			connectClient[raftv1connect.BootstrapServiceClient](&s.ReplicaBootstrapConnMap, &s.ReplicaBootstrapConnMapLock, addr, raftv1connect.NewBootstrapServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr)))
			connectClient[raftv1connect.ElectionServiceClient](&s.ReplicaElectionConnMap, &s.ReplicaElectionConnMapLock, addr, raftv1connect.NewElectionServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr)))
			connectClient[raftv1connect.HeartbeatServiceClient](&s.ReplicaHeartbeatConnMap, &s.ReplicaHeartbeatConnMapLock, addr, raftv1connect.NewHeartbeatServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr)))
			connectClient[raftv1connect.ReplicateOperationServiceClient](&s.ReplicaReplicateConnMap, &s.ReplicaReplicateConnMapLock, addr, raftv1connect.NewReplicateOperationServiceClient(http.DefaultClient, fmt.Sprintf("http://%s", addr)))
		}(addr)
	}
	wg.Wait()
	logger.Info("Bootstrapping completed for server.")
}

// startHeartbeatTimeoutProcess initiates the heartbeat timeout process for a FOLLOWER.
func (s *RaftServer) startHeartbeatTimeoutProcess() error {
	timeoutFunc := func() {
		s.role = ROLE_CANDIDATE // The replica becomes a CANDIDATE to contest in the election
		logger.Info("Role changed to CANDIDATE, requesting votes.")
		votesWon := s.requestVotes()
		totalVotes := 1 + votesWon
		totalCandidates := 1 + len(s.ReplicaHeartbeatConnMap)

		// A candidate wins the election and becomes a leader
		// if it receives more than half of the total votes
		if totalVotes >= totalCandidates/2 {
			logger.Info("Election won, becoming LEADER.")
			// If it wins the election, turn it into a LEADER
			// and start sending heartbeat process
			s.role = ROLE_LEADER
			s.currentTerm++
			s.leaderAddr = s.localAddr

			// If the replica becomes a LEADER, it does not need to listen
			// for heartbeat from other replicas anymore, so stop the
			// heartbeat timeout process
			s.Heartbeat.Stop()

			// The LEADER will send heartbeat to the FOLLOWERS
			go s.sendHeartbeatPeriodically()
		} else {
			logger.Info("Election lost, reverting to FOLLOWER.")
			// If it loses the election, turn it back into a FOLLOWER
			s.role = ROLE_FOLLOWER
		}
	}
	// Start/reset heartbeat timeout process for the follower
	// This will trigger the timeoutFunc after a timeout
	if s.role == ROLE_FOLLOWER {
		logger.Info("Starting heartbeat timeout process for FOLLOWER.")
		s.Heartbeat = heartbeat.NewHeartbeat(HEARTBEAT_TIMEOUT, timeoutFunc)
	}
	return nil
}

// requestVotes is called when the heartbeat has timed out
// and the raft server turns into a candidate.
// It returns the number of votes received along with error (if any)
func (s *RaftServer) requestVotes() int {
	var numVotes int = 0

	s.ReplicaElectionConnMapLock.RLock()
	defer s.ReplicaElectionConnMapLock.RUnlock()

	// iterate over replica addresses and request
	// vote from each replica
	for _, conn := range s.ReplicaElectionConnMap {
		response, err := conn.Voting(
			context.Background(),
			connect.NewRequest(&raftv1.VoteRequest{LogfileIndex: uint64(s.commitIndex)}),
		)
		if err != nil {
			logger.Error("error while requesting vote", "error", err)
			return 0
		}
		if response.Msg.VoteType == raftv1.VoteResponse_VOTE_TYPE_GIVEN {
			numVotes += 1
		}
	}
	logger.Info(fmt.Sprintf("Votes received: %d", numVotes)) // Log the number of votes received
	return numVotes
}

// sendHeartbeatPeriodically is called by the leader to
// send a heartbeat to followers every second
func (s *RaftServer) sendHeartbeatPeriodically() {
	// start the process of sending heartbeat for a leader
	for {
		aliveReplicas := s.sendHeartbeat()
		if aliveReplicas < (len(s.ReplicaHeartbeatConnMap)-1)/2 {
			panic("more than half of the replicas are down")
		}
		time.Sleep(HEARTBEAT_PERIOD)
	}
}

func (s *RaftServer) sendHeartbeat() int {
	aliveCount := 0
	s.ReplicaHeartbeatConnMapLock.RLock()
	for _, conn := range s.ReplicaHeartbeatConnMap {
		response, err := conn.SendHeartbeat(
			context.Background(),
			connect.NewRequest(&raftv1.HeartbeatRequest{
				IsAlive: true,
				Addr:    s.leaderAddr,
				Node:    s.BootstrapNodes,
			}),
		)
		if err != nil {
			logger.Error("error while sending heartbeat", "error", err)
		}
		if response != nil && response.Msg.IsAlive {
			aliveCount++
		}
	}
	s.ReplicaHeartbeatConnMapLock.RUnlock()
	logger.Info(fmt.Sprintf("Alive replicas: %d", aliveCount)) // Log the number of alive replicas
	return aliveCount
}

func (s *RaftServer) ConvertToTransaction(operation string) (*logfile.Transaction, error) {
	// structure of operation => operationName:value... for eg: "add:5"
	return &logfile.Transaction{Index: s.commitIndex + 1, Operation: operation, Term: s.currentTerm}, nil
}

// Performs a two phase commit on all the FOLLOWERS
func (s *RaftServer) PerformTwoPhaseCommit(txn *logfile.Transaction) error {
	s.ReplicaReplicateConnMapLock.RLock()
	wg := &sync.WaitGroup{}

	// First phase of the TwoPhaseCommit: Commit operation

	// CommitOperation on self
	if _, err := s.logfile.CommitOperation(s.commitIndex, s.commitIndex, txn); err != nil {
		logger.Error("Failed to commit operation on self", "error", err) // Added logging
		panic(fmt.Errorf(" %v", err))
	}

	logger.Info(fmt.Sprintf("[%s] performing commit operation on %d followers\n", len(s.ReplicaReplicateConnMap)))

	for addr, conn := range s.ReplicaReplicateConnMap {
		logger.Info(fmt.Sprintf("[%s] sending (CommitOperation: %s) to [%s]\n", txn.Operation, addr))
		response, err := conn.CommitOperation(
			context.Background(),
			connect.NewRequest(&raftv1.CommitTransaction{
				ExpectedFinalIndex: int64(s.commitIndex),
				Index:              int64(txn.Index),
				Operation:          txn.Operation,
				Term:               int64(txn.Term),
			}),
		)
		if err != nil {
			logger.Error("error in (CommitOperation)", "error", err)
			// if there is both, an error and a response, the FOLLOWER is missing
			// some logs. So the LEADER will replicate all the missing logs in the FOLLOWER
			if response != nil {
				wg.Add(1)
				go s.replicateMissingLogs(int(response.Msg.LogfileFinalIndex), addr, conn, wg)
			} else {
				logger.Error("No response received while committing operation", "address", addr) // Added logging
				return err
			}
		} else {
			logger.Info(fmt.Sprintf("Successfully sent commit operation to [%s]", addr)) // Added logging
		}
	}

	// wait for all FOLLOWERS to be consistent
	wg.Wait()

	logger.Info(fmt.Sprintf("[%s] performing (ApplyOperation) on %d followers\n", len(s.ReplicaReplicateConnMap)))

	// Second phase of the TwoPhaseCommit: Apply operation

	// ApplyOperation on self
	if _, err := s.logfile.ApplyOperation(); err != nil {
		logger.Error("Failed to apply operation on self", "error", err) // Added logging
		panic(err)
	}

	for _, conn := range s.ReplicaReplicateConnMap {
		_, err := conn.ApplyOperation(
			context.Background(),
			connect.NewRequest(&raftv1.ApplyOperationRequest{}),
		)
		if err != nil {
			logger.Error("error applying operation to follower", "error", err) // Added logging
			return err
		}
	}
	s.ReplicaReplicateConnMapLock.RUnlock()

	s.commitIndex++ // increment the final commitIndex after applying changes

	s.applyCh <- txn

	return nil
}

// `replicateMissingLogs` makes a FOLLOWER consistent with the leader. This is
// called when the FOLLOWER is missing some logs and refuses a commit operation
// request from the LEADER
func (s *RaftServer) replicateMissingLogs(startIndex int, addr string, client raftv1connect.ReplicateOperationServiceClient, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info(fmt.Sprintf("Starting replication of missing logs from index %d to [%s]", startIndex, addr))
	for {
		startIndex++
		txn, err := s.logfile.GetTransactionWithIndex(startIndex)
		if err != nil {
			logger.Error("error fetching (index) from Logfile", "index", startIndex)
			break
		}
		if txn == nil {
			break
		}
		_, err = client.CommitOperation(
			context.Background(),
			connect.NewRequest(&raftv1.CommitTransaction{
				ExpectedFinalIndex: int64(startIndex),
				Index:              int64(txn.Index),
				Operation:          txn.Operation,
				Term:               int64(txn.Term),
			}),
		)
		if err != nil {
			logger.Error("error replicating missing log (index) to [%s]", "index", startIndex, "addr", addr)
		}
	}
}

// Performs the operation requested by the client.
func (s *RaftServer) Apply(operation string) error {
	logger.Info(fmt.Sprintf("[%s] received operation (%s)\n", operation))
	// only the LEADER is allowed to perform the operation
	// and it then replicates that operation across all the nodes.
	// if the current node is not a LEADER, the operation request
	// will be forwarded to the LEADER, who will then perform the operation
	if s.role == ROLE_LEADER {
		txn, err := s.ConvertToTransaction(operation)
		if err != nil {
			return fmt.Errorf("[%s] error while converting to transaction")
		}
		return s.PerformTwoPhaseCommit(txn)
	}
	logger.Info(fmt.Sprintf("[%s] forwarding operation (%s) to leader [%s]\n", operation, s.leaderAddr))
	s.ReplicaReplicateConnMapLock.RLock()
	defer s.ReplicaReplicateConnMapLock.RUnlock()

	// sending operation to the LEADER to perform a TwoPhaseCommit
	return sendOperationToLeader(operation, s.ReplicaReplicateConnMap[s.leaderAddr])
}

// `sendOperationToLeader` is called when an operation reaches a FOLLOWER.
// This function forwards the operation to the LEADER
func sendOperationToLeader(operation string, conn raftv1connect.ReplicateOperationServiceClient) error {
	_, err := conn.ForwardOperation(
		context.Background(),
		connect.NewRequest(&raftv1.ForwardOperationRequest{Operation: operation}),
	)
	if err != nil {
		return err
	}
	return nil
}
