package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/finaptica/sso/internal/grpc/auth"
	auth "github.com/finaptica/sso/internal/services/auth"
	"google.golang.org/grpc"
)

type AppServer struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(logger *slog.Logger, port int, authService *auth.AuthService) *AppServer {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &AppServer{
		log:        logger,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *AppServer) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *AppServer) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AppServer) Stop() {
	const op = "grcpApp.Stop"

	a.log.With(slog.String("op", op)).Info("Stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
