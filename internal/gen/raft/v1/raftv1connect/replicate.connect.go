// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: raft/v1/replicate.proto

package raftv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	http "net/http"
	v1 "raft/internal/gen/raft/v1"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion0_1_0

const (
	// ReplicateOperationServiceName is the fully-qualified name of the ReplicateOperationService
	// service.
	ReplicateOperationServiceName = "raft.v1.ReplicateOperationService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ReplicateOperationServiceCommitOperationProcedure is the fully-qualified name of the
	// ReplicateOperationService's CommitOperation RPC.
	ReplicateOperationServiceCommitOperationProcedure = "/raft.v1.ReplicateOperationService/CommitOperation"
	// ReplicateOperationServiceApplyOperationProcedure is the fully-qualified name of the
	// ReplicateOperationService's ApplyOperation RPC.
	ReplicateOperationServiceApplyOperationProcedure = "/raft.v1.ReplicateOperationService/ApplyOperation"
	// ReplicateOperationServiceForwardOperationProcedure is the fully-qualified name of the
	// ReplicateOperationService's ForwardOperation RPC.
	ReplicateOperationServiceForwardOperationProcedure = "/raft.v1.ReplicateOperationService/ForwardOperation"
)

// ReplicateOperationServiceClient is a client for the raft.v1.ReplicateOperationService service.
type ReplicateOperationServiceClient interface {
	CommitOperation(context.Context, *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error)
	ApplyOperation(context.Context, *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error)
	ForwardOperation(context.Context, *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error)
}

// NewReplicateOperationServiceClient constructs a client for the raft.v1.ReplicateOperationService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewReplicateOperationServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ReplicateOperationServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &replicateOperationServiceClient{
		commitOperation: connect.NewClient[v1.CommitTransaction, v1.CommitOperationResponse](
			httpClient,
			baseURL+ReplicateOperationServiceCommitOperationProcedure,
			opts...,
		),
		applyOperation: connect.NewClient[v1.ApplyOperationRequest, v1.ApplyOperationResponse](
			httpClient,
			baseURL+ReplicateOperationServiceApplyOperationProcedure,
			opts...,
		),
		forwardOperation: connect.NewClient[v1.ForwardOperationRequest, v1.ForwardOperationResponse](
			httpClient,
			baseURL+ReplicateOperationServiceForwardOperationProcedure,
			opts...,
		),
	}
}

// replicateOperationServiceClient implements ReplicateOperationServiceClient.
type replicateOperationServiceClient struct {
	commitOperation  *connect.Client[v1.CommitTransaction, v1.CommitOperationResponse]
	applyOperation   *connect.Client[v1.ApplyOperationRequest, v1.ApplyOperationResponse]
	forwardOperation *connect.Client[v1.ForwardOperationRequest, v1.ForwardOperationResponse]
}

// CommitOperation calls raft.v1.ReplicateOperationService.CommitOperation.
func (c *replicateOperationServiceClient) CommitOperation(ctx context.Context, req *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error) {
	return c.commitOperation.CallUnary(ctx, req)
}

// ApplyOperation calls raft.v1.ReplicateOperationService.ApplyOperation.
func (c *replicateOperationServiceClient) ApplyOperation(ctx context.Context, req *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error) {
	return c.applyOperation.CallUnary(ctx, req)
}

// ForwardOperation calls raft.v1.ReplicateOperationService.ForwardOperation.
func (c *replicateOperationServiceClient) ForwardOperation(ctx context.Context, req *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error) {
	return c.forwardOperation.CallUnary(ctx, req)
}

// ReplicateOperationServiceHandler is an implementation of the raft.v1.ReplicateOperationService
// service.
type ReplicateOperationServiceHandler interface {
	CommitOperation(context.Context, *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error)
	ApplyOperation(context.Context, *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error)
	ForwardOperation(context.Context, *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error)
}

// NewReplicateOperationServiceHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewReplicateOperationServiceHandler(svc ReplicateOperationServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	replicateOperationServiceCommitOperationHandler := connect.NewUnaryHandler(
		ReplicateOperationServiceCommitOperationProcedure,
		svc.CommitOperation,
		opts...,
	)
	replicateOperationServiceApplyOperationHandler := connect.NewUnaryHandler(
		ReplicateOperationServiceApplyOperationProcedure,
		svc.ApplyOperation,
		opts...,
	)
	replicateOperationServiceForwardOperationHandler := connect.NewUnaryHandler(
		ReplicateOperationServiceForwardOperationProcedure,
		svc.ForwardOperation,
		opts...,
	)
	return "/raft.v1.ReplicateOperationService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ReplicateOperationServiceCommitOperationProcedure:
			replicateOperationServiceCommitOperationHandler.ServeHTTP(w, r)
		case ReplicateOperationServiceApplyOperationProcedure:
			replicateOperationServiceApplyOperationHandler.ServeHTTP(w, r)
		case ReplicateOperationServiceForwardOperationProcedure:
			replicateOperationServiceForwardOperationHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedReplicateOperationServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedReplicateOperationServiceHandler struct{}

func (UnimplementedReplicateOperationServiceHandler) CommitOperation(context.Context, *connect.Request[v1.CommitTransaction]) (*connect.Response[v1.CommitOperationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("raft.v1.ReplicateOperationService.CommitOperation is not implemented"))
}

func (UnimplementedReplicateOperationServiceHandler) ApplyOperation(context.Context, *connect.Request[v1.ApplyOperationRequest]) (*connect.Response[v1.ApplyOperationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("raft.v1.ReplicateOperationService.ApplyOperation is not implemented"))
}

func (UnimplementedReplicateOperationServiceHandler) ForwardOperation(context.Context, *connect.Request[v1.ForwardOperationRequest]) (*connect.Response[v1.ForwardOperationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("raft.v1.ReplicateOperationService.ForwardOperation is not implemented"))
}
