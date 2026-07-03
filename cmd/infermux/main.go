package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/greynewell/infermux"
	"github.com/greynewell/infermux/grpcserver"
	"github.com/greynewell/mist-go/cli"
	"github.com/greynewell/mist-go/tokentrace"
)

func main() {
	app := cli.NewApp("infermux", "0.1.0")

	serve := &cli.Command{
		Name:  "serve",
		Usage: "Start the InferMux inference router",
	}
	serve.AddStringFlag("addr", ":8600", "Listen address")
	serve.AddStringFlag("tokentrace", "", "TokenTrace URL for span reporting")
	serve.Run = func(cmd *cli.Command, args []string) error {
		reg := infermux.NewRegistry()
		reg.Register(infermux.NewEchoProvider("echo", []string{"echo-v1"}, time.Millisecond))

		reporter := tokentrace.NewReporter("infermux", cmd.GetString("tokentrace"))
		router := infermux.NewRouter(reg, reporter)
		handler := infermux.NewHandler(router, reg)

		mux := http.NewServeMux()
		mux.HandleFunc("POST /mist", handler.Ingest)
		mux.HandleFunc("POST /infer", handler.InferDirect)
		mux.HandleFunc("GET /providers", handler.Providers)

		addr := cmd.GetString("addr")
		fmt.Printf("infermux listening on %s\n", addr)
		return http.ListenAndServe(addr, mux)
	}
	app.AddCommand(serve)

	serveGRPC := &cli.Command{
		Name:  "serve-grpc",
		Usage: "Start the InferMux gRPC server",
	}
	serveGRPC.AddStringFlag("addr", ":8601", "gRPC listen address")
	serveGRPC.AddStringFlag("tokentrace", "", "TokenTrace URL for span reporting")
	serveGRPC.Run = func(cmd *cli.Command, args []string) error {
		reg := infermux.NewRegistry()
		reg.Register(infermux.NewEchoProvider("echo", []string{"echo-v1"}, time.Millisecond))

		reporter := tokentrace.NewReporter("infermux", cmd.GetString("tokentrace"))
		router := infermux.NewRouter(reg, reporter)
		svc := grpcserver.NewService(router, reg)

		srv, err := grpcserver.New(svc, grpcserver.Options{Addr: cmd.GetString("addr")})
		if err != nil {
			return err
		}

		// Graceful shutdown: flip health to NOT_SERVING, drain in-flight
		// RPCs, then stop; hard-stop after 10s.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			srv.Shutdown(ctx)
		}()

		fmt.Printf("infermux gRPC listening on %s\n", srv.Addr())
		return srv.Serve()
	}
	app.AddCommand(serveGRPC)

	infer := &cli.Command{
		Name:  "infer",
		Usage: "Run a one-shot inference request",
	}
	infer.AddStringFlag("model", "auto", "Model to use")
	infer.AddStringFlag("prompt", "", "Prompt text")
	infer.Run = func(cmd *cli.Command, args []string) error {
		prompt := cmd.GetString("prompt")
		if prompt == "" && len(args) > 0 {
			prompt = args[0]
		}
		if prompt == "" {
			return fmt.Errorf("--prompt is required")
		}

		reg := infermux.NewRegistry()
		reg.Register(infermux.NewEchoProvider("echo", []string{"echo-v1"}, time.Millisecond))
		reporter := tokentrace.NewReporter("infermux", "")
		router := infermux.NewRouter(reg, reporter)

		resp, err := infermux.InferFromCLI(context.Background(), router, cmd.GetString("model"), prompt)
		if err != nil {
			return err
		}
		fmt.Println(resp.Content)
		return nil
	}
	app.AddCommand(infer)

	if err := app.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
