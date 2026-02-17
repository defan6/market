package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	tokenSecret []byte,
	issuer string,
	tokenTTL time.Duration,
) *App {
	grpcApp := grpcapp.New(log, grpcPort, tokenSecret, issuer, tokenTTL)
	return &App{
		GRPCSrv: grpcApp,
	}
}
