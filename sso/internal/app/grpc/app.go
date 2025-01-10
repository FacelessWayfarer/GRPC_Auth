package grpcapp

import (
	"fmt"
	authgrpc "grpcAuthentication/internal/grpc/auth"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

// New creates new gRPC server app
func New(log *slog.Logger, port int, service authgrpc.Auth) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, service)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server
func (a *App) Run() error {
	const mark = "grpcapp.Run"

	log := a.log.With(slog.String("mark", mark), slog.Int("port", a.port))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", mark, err)
	}

	log.Info("gRPC server running on ...", slog.String("addr", listener.Addr().String()))

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", mark, err)
	}
	return nil
}

// Stop stops gRPC server
func (a *App) Stop() {
	const mark = "grpcapp.Stop"

	a.log.With(slog.String("mark", mark)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
