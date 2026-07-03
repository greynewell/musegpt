// Package integration verifies the InferMux gRPC server and client
// end-to-end over a real TCP listener: no mocked transport, the same
// wire path production traffic takes.
package integration

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/greynewell/infermux"
	"github.com/greynewell/infermux/grpcclient"
	"github.com/greynewell/infermux/grpcserver"
	"github.com/greynewell/mist-go/protocol"
	"github.com/greynewell/mist-go/tokentrace"
)

// flakyProvider fails a fixed number of times, then succeeds. Used to
// verify the client's UNAVAILABLE retry policy end-to-end.
type flakyProvider struct {
	mu        sync.Mutex
	failures  int
	callCount int
}

func (f *flakyProvider) Name() string     { return "flaky" }
func (f *flakyProvider) Models() []string { return []string{"flaky-v1"} }

func (f *flakyProvider) Infer(ctx context.Context, req protocol.InferRequest) (protocol.InferResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	if f.callCount <= f.failures {
		return protocol.InferResponse{}, errors.New("simulated upstream outage")
	}
	return protocol.InferResponse{
		Model:        "flaky-v1",
		Provider:     "flaky",
		Content:      "recovered",
		TokensIn:     1,
		TokensOut:    1,
		FinishReason: "stop",
	}, nil
}

// slowProvider blocks until the context is done. Used to verify deadline
// propagation from client through server to provider.
type slowProvider struct{}

func (s *slowProvider) Name() string     { return "slow" }
func (s *slowProvider) Models() []string { return []string{"slow-v1"} }

func (s *slowProvider) Infer(ctx context.Context, req protocol.InferRequest) (protocol.InferResponse, error) {
	<-ctx.Done()
	return protocol.InferResponse{}, ctx.Err()
}

// startServer boots a real gRPC server on a random port and returns a
// connected client. Both are torn down with the test.
func startServer(t *testing.T, extra ...infermux.Provider) *grpcclient.Client {
	t.Helper()

	reg := infermux.NewRegistry()
	reg.Register(infermux.NewEchoProvider("echo", []string{"echo-v1"}, time.Millisecond))
	for _, p := range extra {
		reg.Register(p)
	}
	router := infermux.NewRouter(reg, tokentrace.NewReporter("infermux-test", ""))
	svc := grpcserver.NewService(router, reg)

	srv, err := grpcserver.New(svc, grpcserver.Options{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	go func() { _ = srv.Serve() }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	})

	client, err := grpcclient.New(srv.Addr())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestInferRoundTrip(t *testing.T) {
	client := startServer(t)

	res, err := client.Prompt(context.Background(), "echo-v1", "hello railway")
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	if res.Content != "echo: hello railway" {
		t.Errorf("content = %q, want %q", res.Content, "echo: hello railway")
	}
	if res.Provider != "echo" {
		t.Errorf("provider = %q, want echo", res.Provider)
	}
	if res.TokensIn <= 0 || res.TokensOut <= 0 {
		t.Errorf("token accounting missing: in=%d out=%d", res.TokensIn, res.TokensOut)
	}
	if res.CostUSD <= 0 {
		t.Errorf("cost accounting missing: %v", res.CostUSD)
	}
}

func TestUnknownModelReturnsNotFound(t *testing.T) {
	client := startServer(t)

	_, err := client.Prompt(context.Background(), "no-such-model", "hi")
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound (err=%v)", status.Code(err), err)
	}
}

func TestEmptyMessagesReturnsInvalidArgument(t *testing.T) {
	client := startServer(t)

	_, err := client.Infer(context.Background(), "echo-v1", nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument (err=%v)", status.Code(err), err)
	}
}

func TestInvalidRoleReturnsInvalidArgument(t *testing.T) {
	client := startServer(t)

	_, err := client.Infer(context.Background(), "echo-v1", []grpcclient.Message{
		{Role: "wizard", Content: "cast fireball"},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument (err=%v)", status.Code(err), err)
	}
}

func TestClientRetriesUnavailableThenSucceeds(t *testing.T) {
	flaky := &flakyProvider{failures: 2}
	client := startServer(t, flaky)

	res, err := client.Prompt(context.Background(), "flaky-v1", "are you up")
	if err != nil {
		t.Fatalf("expected retry to recover, got %v", err)
	}
	if res.Content != "recovered" {
		t.Errorf("content = %q, want recovered", res.Content)
	}
	if flaky.callCount != 3 {
		t.Errorf("provider called %d times, want 3 (2 failures + 1 success)", flaky.callCount)
	}
}

func TestRetriesExhaustedSurfacesUnavailable(t *testing.T) {
	flaky := &flakyProvider{failures: 100}
	client := startServer(t, flaky)

	_, err := client.Prompt(context.Background(), "flaky-v1", "hi")
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("code = %v, want Unavailable (err=%v)", status.Code(err), err)
	}
	if flaky.callCount != 3 {
		t.Errorf("provider called %d times, want exactly maxAttempts=3", flaky.callCount)
	}
}

func TestDeadlinePropagatesToProvider(t *testing.T) {
	client := startServer(t, &slowProvider{})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Prompt(ctx, "slow-v1", "hang forever")
	elapsed := time.Since(start)

	if code := status.Code(err); code != codes.DeadlineExceeded {
		t.Fatalf("code = %v, want DeadlineExceeded (err=%v)", code, err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("deadline not propagated: call took %v", elapsed)
	}
}

func TestListProviders(t *testing.T) {
	client := startServer(t)

	providers, err := client.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}
	if len(providers) != 1 || providers[0].Name != "echo" {
		t.Fatalf("providers = %+v, want [echo]", providers)
	}
	if len(providers[0].Models) != 1 || providers[0].Models[0] != "echo-v1" {
		t.Errorf("models = %v, want [echo-v1]", providers[0].Models)
	}
}

func TestHealthCheckServing(t *testing.T) {
	reg := infermux.NewRegistry()
	reg.Register(infermux.NewEchoProvider("echo", []string{"echo-v1"}, time.Millisecond))
	router := infermux.NewRouter(reg, tokentrace.NewReporter("infermux-test", ""))
	srv, err := grpcserver.New(grpcserver.NewService(router, reg), grpcserver.Options{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	go func() { _ = srv.Serve() }()

	// Probe the standard health service the way a load balancer would.
	hc, err := grpcclient.New(srv.Addr())
	if err != nil {
		t.Fatalf("dial health: %v", err)
	}
	defer hc.Close()
	conn := hc.Conn()
	resp, err := healthpb.NewHealthClient(conn).Check(context.Background(), &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("health = %v, want SERVING", resp.GetStatus())
	}

	// Graceful shutdown flips health to NOT_SERVING and completes.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
