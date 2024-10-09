package raft

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"raft/internal/fs"
	raftv1 "raft/internal/gen/raft/v1"
	"raft/internal/gen/raft/v1/raftv1connect"
	"raft/internal/heartbeat"
	"raft/internal/logfile"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
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
	Address        string                    // Address of the server
	role           ROLE                      // Role of the server
	BootstrapNodes []string                  // List of bootstrap nodes for the cluster
	Heartbeat      *heartbeat.Heartbeat      // Heartbeat service
	logfile        logfile.Log               // Log file for the server
	applyCh        chan *logfile.Transaction // Channel for applying transactions
}

// RaftServer represents a Raft server instance.
type RaftServer struct {
	role           ROLE                      // Current role of the server
	BootstrapNodes []string                  // List of bootstrap nodes
	Heartbeat      *heartbeat.Heartbeat      // Heartbeat service
	logfile        logfile.Log               // Log file for the server
	applyCh        chan *logfile.Transaction // Channel for applying transactions
	Transport      *http.Server              // HTTP server for transport
	leaderAddr     string                    // Address of the current leader
	currentTerm    int                       // Current term of the server
	commitIndex    int                       // Index of the last committed entry

	// ReplicaConnMap maintains connections to other replicas to reduce latency.
	ReplicaConnMap     map[string]*grpc.ClientConn
	ReplicaConnMapLock sync.RWMutex
}

// NewRaftServer creates a new Raft server instance with the provided options.
func NewRaftServer(opts RaftServerOpts, srv *http.Server) (*RaftServer, map[string]string) {
	raftServer := &RaftServer{
		role:           opts.role,
		BootstrapNodes: opts.BootstrapNodes,
		Heartbeat:      opts.Heartbeat,
		Transport:      srv,
		logfile:        opts.logfile,
		applyCh:        opts.applyCh,
	}
	filePath := fmt.Sprintf("%s/%s.%s", SNAPSHOTS_DIR, opts.Address, fs.FILE_EXTENSION)
	_, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[%s] Snapshot not found, creating a new RaftServer instance.", opts.Address)
		return raftServer, make(map[string]string)
	}
	log.Printf("[%s] Restoring RaftServer from snapshot.", opts.Address)

	// If a snapshot exists, restore additional configurations to the server.
	snapshotContent, err := fs.ReadFile(SNAPSHOTS_DIR, opts.Address)
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

	log.Println("Raft server started successfully.")
	return nil
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
			if _, err := client.SendHeartbeat(context.Background(), connect.NewRequest(&raftv1.HeartbeatRequest{
				SenderId: "",
				Term:     0,
			})); err != nil {
				log.Printf("Failed to connect to [%s]: %v", addr, err)
			} else {
				log.Printf("Successfully connected to [%s].", addr)
			}
			wg.Done()
		}(s, addr, wg)
	}
	wg.Wait()
	log.Printf("Bootstrapping completed for server [%s].", s.Transport.Addr)
}
