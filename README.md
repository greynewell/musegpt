# infermux

Inference router. Part of the [MIST stack](https://github.com/greynewell/mist-go).

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Install

```bash
go get github.com/greynewell/infermux
```

## Provider interface

```go
type Provider interface {
    Name() string
    Models() []string
    Infer(ctx context.Context, req protocol.InferRequest) (protocol.InferResponse, error)
}
```

## Route

```go
reg := infermux.NewRegistry()
reg.Register(myOpenAIProvider)
reg.Register(myAnthropicProvider)

reporter := tokentrace.NewReporter("infermux", "http://localhost:8700")
router := infermux.NewRouter(reg, reporter)

resp, err := router.Infer(ctx, protocol.InferRequest{
    Model:    "claude-sonnet-4-5-20250929",
    Messages: []protocol.ChatMessage{{Role: "user", Content: "Hello"}},
})
```

Tracks tokens and cost per request. Reports spans to [TokenTrace](https://github.com/greynewell/tokentrace).

## HTTP API

```go
handler := infermux.NewHandler(router, reg)
http.HandleFunc("POST /mist", handler.Ingest)
http.HandleFunc("POST /infer", handler.InferDirect)
http.HandleFunc("GET /providers", handler.Providers)
```

## gRPC API

InferMux serves the same router over gRPC (`infermux.v1.InferMuxService`), defined in [`proto/infermux/v1/infermux.proto`](proto/infermux/v1/infermux.proto).

```bash
infermux serve-grpc --addr :8601 --tokentrace http://localhost:8700
```

The server ships with the standard gRPC health service, server reflection (works with `grpcurl` out of the box), keepalive enforcement, panic recovery, structured per-RPC logging, and graceful drain on SIGINT/SIGTERM.

Go client:

```go
import "github.com/greynewell/infermux/grpcclient"

c, err := grpcclient.New("localhost:8601")
defer c.Close()

res, err := c.Prompt(ctx, "echo-v1", "Hello world")
// res.Content, res.Provider, res.TokensIn, res.TokensOut, res.CostUSD
```

The client retries `UNAVAILABLE` (transient provider failure) up to 3 attempts with exponential backoff via gRPC service config, and never retries `NOT_FOUND` or `INVALID_ARGUMENT`. Caller deadlines propagate through the server into provider calls.

Error contract:

| Condition | gRPC code |
|---|---|
| Empty messages, bad role, temperature out of range | `INVALID_ARGUMENT` |
| No provider for the requested model | `NOT_FOUND` |
| Resolved provider failed upstream (retryable) | `UNAVAILABLE` |
| Caller deadline elapsed | `DEADLINE_EXCEEDED` |

Integration tests cover the full wire path (real TCP, real server, real client), including retry behavior, deadline propagation, error mapping, and health checks:

```bash
go test ./integration/ -race -v
```

Regenerate protobuf stubs:

```bash
protoc --proto_path=proto \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  proto/infermux/v1/infermux.proto
```

## CLI

```bash
infermux serve --addr :8600 --tokentrace http://localhost:8700
infermux serve-grpc --addr :8601 --tokentrace http://localhost:8700
infermux infer --model echo-v1 --prompt "Hello world"
```
