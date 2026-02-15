package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"
	"sso/internal/grpc/auth/middleware"
	"sso/internal/lib/security/encoder"
	"sso/internal/lib/security/token/generator"
	"sso/internal/lib/security/token/signer"
	"sso/internal/service"
	"sso/internal/storage"
	db "sso/storage"
	"time"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
	tokenSecret []byte,
	issuer string,
	tokenTTL time.Duration,
) *App {

	database := db.NewDatabase()
	storer := storage.NewStorage(database.GetDB(), log)
	passwordEncoder := encoder.NewPasswordEncoder()
	tokenSigner := signer.NewHMACSigner(tokenSecret)
	tokenGenerator := generator.NewDefaultTokenGenerator(tokenSigner, issuer, tokenTTL)
	authService := service.NewDefaultAuthService(log, storer, storer, passwordEncoder, tokenGenerator)
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(
		middleware.AuthInterceptor(tokenSigner)))
	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
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
	if err := a.GRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.GRPCServer.GracefulStop()
}
