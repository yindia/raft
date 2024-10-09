// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: raft/v1/heartbeat.proto

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
	// HeartbeatServiceName is the fully-qualified name of the HeartbeatService service.
	HeartbeatServiceName = "raft.v1.HeartbeatService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// HeartbeatServiceSendHeartbeatProcedure is the fully-qualified name of the HeartbeatService's
	// SendHeartbeat RPC.
	HeartbeatServiceSendHeartbeatProcedure = "/raft.v1.HeartbeatService/SendHeartbeat"
)

// HeartbeatServiceClient is a client for the raft.v1.HeartbeatService service.
type HeartbeatServiceClient interface {
	// SendHeartbeat handles heartbeat requests from nodes.
	SendHeartbeat(context.Context, *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error)
}

// NewHeartbeatServiceClient constructs a client for the raft.v1.HeartbeatService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewHeartbeatServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) HeartbeatServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &heartbeatServiceClient{
		sendHeartbeat: connect.NewClient[v1.HeartbeatRequest, v1.HeartbeatResponse](
			httpClient,
			baseURL+HeartbeatServiceSendHeartbeatProcedure,
			opts...,
		),
	}
}

// heartbeatServiceClient implements HeartbeatServiceClient.
type heartbeatServiceClient struct {
	sendHeartbeat *connect.Client[v1.HeartbeatRequest, v1.HeartbeatResponse]
}

// SendHeartbeat calls raft.v1.HeartbeatService.SendHeartbeat.
func (c *heartbeatServiceClient) SendHeartbeat(ctx context.Context, req *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error) {
	return c.sendHeartbeat.CallUnary(ctx, req)
}

// HeartbeatServiceHandler is an implementation of the raft.v1.HeartbeatService service.
type HeartbeatServiceHandler interface {
	// SendHeartbeat handles heartbeat requests from nodes.
	SendHeartbeat(context.Context, *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error)
}

// NewHeartbeatServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewHeartbeatServiceHandler(svc HeartbeatServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	heartbeatServiceSendHeartbeatHandler := connect.NewUnaryHandler(
		HeartbeatServiceSendHeartbeatProcedure,
		svc.SendHeartbeat,
		opts...,
	)
	return "/raft.v1.HeartbeatService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case HeartbeatServiceSendHeartbeatProcedure:
			heartbeatServiceSendHeartbeatHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedHeartbeatServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedHeartbeatServiceHandler struct{}

func (UnimplementedHeartbeatServiceHandler) SendHeartbeat(context.Context, *connect.Request[v1.HeartbeatRequest]) (*connect.Response[v1.HeartbeatResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("raft.v1.HeartbeatService.SendHeartbeat is not implemented"))
}
