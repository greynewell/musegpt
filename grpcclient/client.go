// Package grpcclient is a thin, ergonomic Go client for the InferMux
// gRPC API.
//
// Design decisions:
//
//   - The client speaks domain types (roles, prompts) and hides protobuf
//     plumbing from callers.
//   - Every call requires a context; the client applies a default timeout
//     only when the caller has not set a deadline, so explicit deadlines
//     always win.
//   - Retries are declared in a client-side service config for UNAVAILABLE
//     only. Infer is not idempotent in general (real providers bill per
//     call), so retry count is kept low and NOT_FOUND / INVALID_ARGUMENT
//     are never retried.
package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	infermuxv1 "github.com/greynewell/infermux/gen/infermux/v1"
)

// retryPolicy retries only UNAVAILABLE (the code the server reserves for
// transient upstream provider failure) with capped exponential backoff.
const retryPolicy = `{
  "methodConfig": [{
    "name": [{"service": "infermux.v1.InferMuxService"}],
    "retryPolicy": {
      "maxAttempts": 3,
      "initialBackoff": "0.1s",
      "maxBackoff": "1s",
      "backoffMultiplier": 2,
      "retryableStatusCodes": ["UNAVAILABLE"]
    }
  }]
}`

// Client is an InferMux gRPC client.
type Client struct {
	conn           *grpc.ClientConn
	svc            infermuxv1.InferMuxServiceClient
	defaultTimeout time.Duration
}

// Option customizes the client.
type Option func(*config)

type config struct {
	defaultTimeout time.Duration
	dialOpts       []grpc.DialOption
}

// WithDefaultTimeout sets the timeout applied when the caller's context
// has no deadline. Defaults to 30s.
func WithDefaultTimeout(d time.Duration) Option {
	return func(c *config) { c.defaultTimeout = d }
}

// WithDialOptions appends extra grpc.DialOptions (e.g. TLS credentials).
func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(c *config) { c.dialOpts = append(c.dialOpts, opts...) }
}

// New connects to an InferMux gRPC server at target (host:port).
//
// Transport security defaults to insecure credentials for local and
// private-network use; production callers pass TLS credentials through
// WithDialOptions, which take precedence by ordering.
func New(target string, opts ...Option) (*Client, error) {
	cfg := &config{defaultTimeout: 30 * time.Second}
	for _, o := range opts {
		o(cfg)
	}

	dialOpts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	}, cfg.dialOpts...)

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:           conn,
		svc:            infermuxv1.NewInferMuxServiceClient(conn),
		defaultTimeout: cfg.defaultTimeout,
	}, nil
}

// Close tears down the underlying connection.
func (c *Client) Close() error { return c.conn.Close() }

// Conn exposes the underlying connection so callers can attach other
// gRPC services (e.g. the standard health client) to the same channel.
func (c *Client) Conn() *grpc.ClientConn { return c.conn }

// Message is one conversation turn.
type Message struct {
	Role    string
	Content string
}

// InferResult is the completed inference with routing metadata.
type InferResult struct {
	Model        string
	Provider     string
	Content      string
	TokensIn     int64
	TokensOut    int64
	CostUSD      float64
	LatencyMS    int64
	FinishReason string
}

// Infer sends a conversation to the router and returns the response.
func (c *Client) Infer(ctx context.Context, model string, messages []Message) (*InferResult, error) {
	ctx, cancel := c.withDefaultDeadline(ctx)
	defer cancel()

	req := &infermuxv1.InferRequest{Model: model}
	for _, m := range messages {
		req.Messages = append(req.Messages, &infermuxv1.ChatMessage{Role: m.Role, Content: m.Content})
	}

	resp, err := c.svc.Infer(ctx, req)
	if err != nil {
		return nil, err
	}
	return &InferResult{
		Model:        resp.GetModel(),
		Provider:     resp.GetProvider(),
		Content:      resp.GetContent(),
		TokensIn:     resp.GetTokensIn(),
		TokensOut:    resp.GetTokensOut(),
		CostUSD:      resp.GetCostUsd(),
		LatencyMS:    resp.GetLatencyMs(),
		FinishReason: resp.GetFinishReason(),
	}, nil
}

// Prompt is a convenience wrapper for a single user message.
func (c *Client) Prompt(ctx context.Context, model, prompt string) (*InferResult, error) {
	return c.Infer(ctx, model, []Message{{Role: "user", Content: prompt}})
}

// ProviderInfo describes one registered provider.
type ProviderInfo struct {
	Name   string
	Models []string
}

// ListProviders returns all providers registered on the server.
func (c *Client) ListProviders(ctx context.Context) ([]ProviderInfo, error) {
	ctx, cancel := c.withDefaultDeadline(ctx)
	defer cancel()

	resp, err := c.svc.ListProviders(ctx, &infermuxv1.ListProvidersRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]ProviderInfo, 0, len(resp.GetProviders()))
	for _, p := range resp.GetProviders() {
		out = append(out, ProviderInfo{Name: p.GetName(), Models: p.GetModels()})
	}
	return out, nil
}

// withDefaultDeadline applies the default timeout only when the caller
// did not set a deadline, so explicit caller deadlines always win.
func (c *Client) withDefaultDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.defaultTimeout)
}
