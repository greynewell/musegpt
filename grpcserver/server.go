// Package grpcserver exposes the InferMux router over gRPC.
//
// Design decisions:
//
//   - Transport-only layer: this package translates between protobuf types
//     and the core infermux domain types. All routing, tracing, and cost
//     accounting stays in the core package, so HTTP and gRPC serve
//     identical behavior from one code path.
//   - Errors are mapped to canonical gRPC status codes via errors.Is /
//     errors.As, never string matching: unknown model -> NOT_FOUND,
//     bad request -> INVALID_ARGUMENT, provider failure -> UNAVAILABLE,
//     caller deadline -> DEADLINE_EXCEEDED.
//   - The server registers the standard gRPC health service and server
//     reflection, so load balancers can probe readiness and grpcurl can
//     explore the API without the .proto file.
package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/greynewell/infermux"
	infermuxv1 "github.com/greynewell/infermux/gen/infermux/v1"
	"github.com/greynewell/mist-go/protocol"
)

// Service implements infermux.v1.InferMuxService on top of the core router.
type Service struct {
	infermuxv1.UnimplementedInferMuxServiceServer

	router   *infermux.Router
	registry *infermux.Registry
}

// NewService wires a gRPC service to the given router and registry.
func NewService(router *infermux.Router, registry *infermux.Registry) *Service {
	return &Service{router: router, registry: registry}
}

// Infer routes one inference request. The caller's context (and therefore
// its deadline and cancellation) propagates through the router into the
// provider call, so a client-side timeout cancels upstream work instead of
// leaking it.
func (s *Service) Infer(ctx context.Context, req *infermuxv1.InferRequest) (*infermuxv1.InferResponse, error) {
	if len(req.GetMessages()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "messages must not be empty")
	}
	if t := req.GetTemperature(); req.GetTemperatureSet() && (t < 0 || t > 2) {
		return nil, status.Errorf(codes.InvalidArgument, "temperature %v out of range [0, 2]", t)
	}

	domainReq := protocol.InferRequest{
		Model:    req.GetModel(),
		Messages: make([]protocol.ChatMessage, 0, len(req.GetMessages())),
	}
	for _, m := range req.GetMessages() {
		switch m.GetRole() {
		case "system", "user", "assistant":
		default:
			return nil, status.Errorf(codes.InvalidArgument, "invalid role %q", m.GetRole())
		}
		domainReq.Messages = append(domainReq.Messages, protocol.ChatMessage{
			Role:    m.GetRole(),
			Content: m.GetContent(),
		})
	}

	resp, err := s.router.Infer(ctx, domainReq)
	if err != nil {
		return nil, mapError(ctx, err)
	}

	return &infermuxv1.InferResponse{
		Model:        resp.Model,
		Provider:     resp.Provider,
		Content:      resp.Content,
		TokensIn:     resp.TokensIn,
		TokensOut:    resp.TokensOut,
		CostUsd:      resp.CostUSD,
		LatencyMs:    resp.LatencyMS,
		FinishReason: resp.FinishReason,
	}, nil
}

// ListProviders returns every registered provider and its models.
func (s *Service) ListProviders(ctx context.Context, _ *infermuxv1.ListProvidersRequest) (*infermuxv1.ListProvidersResponse, error) {
	out := &infermuxv1.ListProvidersResponse{}
	for _, name := range s.registry.Providers() {
		if p, ok := s.registry.Get(name); ok {
			out.Providers = append(out.Providers, &infermuxv1.ProviderInfo{
				Name:   p.Name(),
				Models: p.Models(),
			})
		}
	}
	return out, nil
}

// mapError converts core errors into canonical gRPC status codes.
func mapError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, infermux.ErrNoProvider):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled:
		return status.Error(codes.Canceled, err.Error())
	default:
		// A resolved provider failed upstream: retryable by contract.
		return status.Error(codes.Unavailable, err.Error())
	}
}

// Server wraps a configured *grpc.Server with its listener and health state.
type Server struct {
	grpcServer *grpc.Server
	health     *health.Server
	lis        net.Listener
}

// Options configures the gRPC server.
type Options struct {
	// Addr is the TCP listen address, e.g. ":8601".
	Addr string
	// MaxRecvMsgSize caps inbound message size in bytes. Zero uses 4 MiB.
	MaxRecvMsgSize int
	// Logger receives one structured line per RPC. Nil uses slog.Default().
	Logger *slog.Logger
}

// New builds a Server with health checks, reflection, keepalive policy,
// panic recovery, and structured per-RPC logging.
func New(svc *Service, opts Options) (*Server, error) {
	if opts.MaxRecvMsgSize == 0 {
		opts.MaxRecvMsgSize = 4 << 20
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	lis, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	gs := grpc.NewServer(
		grpc.MaxRecvMsgSize(opts.MaxRecvMsgSize),
		// Tolerate keepalive pings as fast as every 10s even without
		// active streams; kills silent connection drops behind NATs and
		// L4 load balancers without letting clients ping-flood.
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(logger),
			loggingInterceptor(logger),
		),
	)

	infermuxv1.RegisterInferMuxServiceServer(gs, svc)

	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus(infermuxv1.InferMuxService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gs, hs)

	reflection.Register(gs)

	return &Server{grpcServer: gs, health: hs, lis: lis}, nil
}

// Addr returns the bound listen address (useful with ":0" in tests).
func (s *Server) Addr() string { return s.lis.Addr().String() }

// Serve blocks serving RPCs until Shutdown or Stop is called.
func (s *Server) Serve() error { return s.grpcServer.Serve(s.lis) }

// Shutdown gracefully stops the server: health flips to NOT_SERVING so
// load balancers drain, in-flight RPCs complete, then the server stops.
// If ctx expires first, the server is stopped hard.
func (s *Server) Shutdown(ctx context.Context) {
	s.health.Shutdown()
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		s.grpcServer.Stop()
	}
}

// recoveryInterceptor converts handler panics into codes.Internal instead
// of crashing the process and taking every in-flight RPC down with it.
func recoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("rpc panic", "method", info.FullMethod, "panic", r)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// loggingInterceptor emits one structured log line per RPC with method,
// status code, and wall latency.
func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Info("rpc",
			"method", info.FullMethod,
			"code", status.Code(err).String(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
		return resp, err
	}
}
