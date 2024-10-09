package raft

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	raftv1 "raft/internal/gen/raft/v1"
	"raft/internal/gen/raft/v1/raftv1connect"

	"raft/internal/heartbeat"
	"raft/internal/logfile"
	"sync"
	"time"

	"connectrpc.com/connect"
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
	HEARTBEAT_PERIOD  = time.Second * 1  // Duration between heartbeats
	HEARTBEAT_TIMEOUT = time.Second * 10 // Timeout for heartbeat responses
)

// RaftServerOpts holds the options for creating a new Raft server.
type RaftServerOpts struct {
	Address        string   // Address of the server
	Role           ROLE     // Role of the server
	BootstrapNodes []string // List of bootstrap nodes for the cluster
}

// RaftServer represents a Raft server instance.
type RaftServer struct {
	role           ROLE                      // Current role of the server
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

		ReplicaElectionConnMap:  make(ReplicaConnMap[string, raftv1connect.ElectionServiceClient]),
		ReplicaHeartbeatConnMap: make(ReplicaConnMap[string, raftv1connect.HeartbeatServiceClient]),
		ReplicaBootstrapConnMap: make(ReplicaConnMap[string, raftv1connect.BootstrapServiceClient]),
	}
	filePath := fmt.Sprintf("%s/%s.%s", SNAPSHOTS_DIR, opts.Address, FILE_EXTENSION)
	_, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[%s] Snapshot not found. Creating a new RaftServer instance.", opts.Address)
		return raftServer, make(map[string]string)
	}
	log.Printf("[%s] Restoring RaftServer from snapshot.", opts.Address)

	// If a snapshot exists, restore additional configurations to the server.
	snapshotContent, err := readFile(SNAPSHOTS_DIR, opts.Address)
	if err != nil {
		log.Fatalf("Error while reading snapshot: %v", err)
	}
	var kvMap map[string]string
	raftServer.commitIndex, kvMap = destructureSnapshot(snapshotContent)
	return raftServer, kvMap
}

// Start initializes the Raft server and begins the heartbeat process.
func (s *RaftServer) Start() error {
	time.Sleep(time.Second * 3) // Wait for the server to start

	// Send requests to bootstrapped servers to add this server to their `replicaConnMap`.
	s.bootstrapNetwork()

	s.startHeartbeatTimeoutProcess()

	log.Println("Raft server started successfully.")
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

// bootstrapNetwork sends requests to other replicas to add this server to their replicaConnMap.
func (s *RaftServer) bootstrapNetwork() {
	wg := &sync.WaitGroup{}
	for _, addr := range s.BootstrapNodes {
		wg.Add(1)
		if len(addr) == 0 {
			log.Printf("Skipping empty bootstrap address.")
			wg.Done()
			continue
		}
		go func(s *RaftServer, addr string, wg *sync.WaitGroup) {
			log.Printf("Attempting to connect with bootstrap node [%s].", addr)
			client := raftv1connect.NewHeartbeatServiceClient(http.DefaultClient, addr)
			if _, err := client.SendHeartbeat(context.Background(), connect.NewRequest(&raftv1.HeartbeatRequest{})); err != nil {
				log.Printf("Failed to connect to bootstrap node [%s]: %v", addr, err)
			} else {
				log.Printf("Successfully connected to bootstrap node [%s].", addr)
			}
			wg.Done()
		}(s, addr, wg)
	}
	wg.Wait()
	log.Printf("Bootstrapping completed for server.")
}

// startHeartbeatTimeoutProcess initiates the heartbeat timeout process for a FOLLOWER.
func (s *RaftServer) startHeartbeatTimeoutProcess() error {
	timeoutFunc := func() {
		s.role = ROLE_CANDIDATE // The replica becomes a CANDIDATE to contest in the election
		votesWon := s.requestVotes()
		totalVotes := 1 + votesWon
		totalCandidates := 1 + len(s.ReplicaHeartbeatConnMap)

		// A candidate wins the election and becomes a leader
		// if it receives more than half of the total votes
		if totalVotes >= totalCandidates/2 {
			// If it wins the election, turn it into a LEADER
			// and start sending heartbeat process
			s.role = ROLE_LEADER
			s.currentTerm++
			s.leaderAddr = s.leaderAddr

			// If the replica becomes a LEADER, it does not need to listen
			// for heartbeat from other replicas anymore, so stop the
			// heartbeat timeout process
			s.Heartbeat.Stop()

			// The LEADER will send heartbeat to the FOLLOWERS
			go s.sendHeartbeatPeriodically()
		} else {
			// If it loses the election, turn it back into a FOLLOWER
			s.role = ROLE_FOLLOWER
		}
	}
	// Start/reset heartbeat timeout process for the follower
	// This will trigger the timeoutFunc after a timeout
	if s.role == ROLE_FOLLOWER {
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
			log.Printf("error while requesting vote: %v\n", err)
			return 0
		}
		if response.Msg.VoteType == raftv1.VoteResponse_VOTE_TYPE_GIVEN {
			numVotes += 1
		}
	}
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
			}),
		)
		if err != nil {
			log.Printf("error while sending heartbeat to [%s]: %v\n", err)
		}
		if response != nil && response.Msg.IsAlive {
			aliveCount++
		}
	}
	s.ReplicaHeartbeatConnMapLock.RUnlock()
	return aliveCount
}
