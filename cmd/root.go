package cmd

import (
	"context"
	raftv1 "raft/internal/gen/raft/v1"
	"raft/internal/gen/raft/v1/raftv1connect"
	"raft/internal/raft"

	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"raft/route"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"go.akshayshah.org/connectauth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Configuration variables
var (
	cfgFile   string
	logLevel  string
	httpAddr  string // Changed from 'address' to 'httpAddr' for clarity
	raftDir   string
	raftAddr  string
	joinAddr  string
	nodeID    string
	bootstrap bool
)

const (
	compressMinBytes = 1024 // Minimum byte size for compression
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "redis-cli",
	Short: "A CLI application for managing a distributed Redis-like system",
	Long:  `redis-cli is a command-line interface for managing a distributed Redis-like system using Raft consensus.`,
	Run:   runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute root command", "error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.redis-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")

	// Local flags
	rootCmd.Flags().StringVar(&raftDir, "dir", "", "Raft storage directory")
	rootCmd.Flags().StringVar(&httpAddr, "addr", "127.0.0.1:12000", "HTTP server address")
	rootCmd.Flags().StringVar(&joinAddr, "join", "", "Set join address, if any")
	rootCmd.Flags().StringVar(&nodeID, "id", "", "Node ID. If not set, same as Raft bind address")

}

func initConfig() {
	// TODO: Implement config file reading logic if needed
}

func runRoot(cmd *cobra.Command, args []string) {
	if err := runRootCommand(); err != nil {
		slog.Error("Error running root command", "error", err)
		os.Exit(1)
	}
}

func runRootCommand() error {
	initLogger(logLevel)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	middleware := connectauth.NewMiddleware(authenticateRequest)
	raftServer, _ := raft.NewRaftServer(raft.RaftServerOpts{
		Address:        httpAddr,
		BootstrapNodes: []string{},
		Role:           raft.ROLE_FOLLOWER,
	})
	raftServer.Start()

	mux := http.NewServeMux()
	if err := setupHandlers(mux, middleware, raftServer); err != nil {
		return fmt.Errorf("failed to set up handlers: %w", err)
	}

	mux.Handle("/metrics", promhttp.Handler())

	srv := initializeHTTPServer(mux)

	serverErrChan := startServer(srv)

	if joinAddr != "" {
		fmt.Println(raftv1connect.NewBootstrapServiceClient(http.DefaultClient, joinAddr).AddReplica(context.Background(), connect.NewRequest((&raftv1.AddrInfo{
			Addr: httpAddr,
		}))))

	}
	return handleServerLifecycle(srv, exitChan, serverErrChan)
}

func initLogger(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		slog.Warn("Invalid log level, using 'info' as default", "level", level)
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
}

func authenticateRequest(ctx context.Context, req *connectauth.Request) (any, error) {
	// TODO: Implement your authentication logic here
	return nil, nil
}

func setupHandlers(mux *http.ServeMux, middleware *connectauth.Middleware, raftServer *raft.RaftServer) error {
	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return fmt.Errorf("failed to create interceptor: %w", err)
	}

	pattern, handler := raftv1connect.NewBootstrapServiceHandler(
		route.NewBootstrapServer(raftServer),
		connect.WithInterceptors(otelInterceptor),
		connect.WithCompressMinBytes(compressMinBytes),
	)

	mux.Handle(pattern, middleware.Wrap(handler))

	pattern, handler = raftv1connect.NewElectionServiceHandler(
		route.NewElectionServer(raftServer),
		connect.WithInterceptors(otelInterceptor),
		connect.WithCompressMinBytes(compressMinBytes),
	)

	mux.Handle(pattern, middleware.Wrap(handler))

	pattern, handler = raftv1connect.NewHeartbeatServiceHandler(
		route.NewHeartbeatServer(raftServer),
		connect.WithInterceptors(otelInterceptor),
		connect.WithCompressMinBytes(compressMinBytes),
	)

	mux.Handle(pattern, middleware.Wrap(handler))

	mux.Handle(grpchealth.NewHandler(
		grpchealth.NewStaticChecker(raftv1connect.BootstrapServiceName),
	))
	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(raftv1connect.BootstrapServiceName),
	))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(raftv1connect.BootstrapServiceName),
	))

	slog.Info("Handlers set up successfully", "serviceName", raftv1connect.BootstrapServiceName)
	return nil
}

func initializeHTTPServer(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr: httpAddr, // Use httpAddr instead of raftAddr
		Handler: h2c.NewHandler(
			newCORS().Handler(mux),
			&http2.Server{},
		),
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		MaxHeaderBytes:    8 * 1024, // 8KiB
	}
}

func startServer(srv *http.Server) chan error {
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("HTTP server starting", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()
	return serverErrChan
}

func handleServerLifecycle(srv *http.Server, exitChan chan os.Signal, serverErrChan chan error) error {
	select {
	case <-exitChan:
		slog.Info("Shutdown signal received, shutting down server...")
	case err := <-serverErrChan:
		return err
	}

	if err := shutdownServer(srv); err != nil {
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}
	slog.Info("HTTP server shut down gracefully")
	return nil
}

func shutdownServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func newCORS() *cors.Cors {
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowOriginFunc: func(origin string) bool {
			return true // Allow all origins
		},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{
			"Accept",
			"Accept-Encoding",
			"Accept-Post",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Content-Encoding",
			"Grpc-Accept-Encoding",
			"Grpc-Encoding",
			"Grpc-Message",
			"Grpc-Status",
			"Grpc-Status-Details-Bin",
		},
	})
}
