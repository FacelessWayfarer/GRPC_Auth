package app

import (
	grpcapp "grpcAuthentication/internal/app/grpc"
	sqllite "grpcAuthentication/internal/database/sqlite"
	"grpcAuthentication/internal/services/auth"
	"log/slog"
	"time"
)

type App struct {
	GRPCServ *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	database, err := sqllite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, database, database, database, tokenTTL)

	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{GRPCServ: grpcApp}
}
