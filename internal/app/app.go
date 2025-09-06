package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/finaptica/sso/internal/app/grpc"
	"github.com/finaptica/sso/internal/services"
	"github.com/finaptica/sso/internal/storage"
	"github.com/finaptica/sso/internal/storage/repository"
)

type App struct {
	GRPCSrv *grpcapp.AppServer
}

func New(log *slog.Logger, grpcPort int, postgresConnectionString string, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *App {
	db, err := storage.New(log, postgresConnectionString)
	if err != nil {
		log.Error("failed to init storage", slog.String("err", err.Error()))
		panic(err)
	}
	userRepository := repository.NewUserRepository(log, db)
	appRepository := repository.NewAppRepository(log, db)
	refreshTokenRepository := repository.NewRefreshTokenRepository(log, db)
	authService := services.NewAuthService(log, userRepository, appRepository, refreshTokenRepository, accessTokenTTL, refreshTokenTTL)
	rtsService := services.NewRefreshTokenService(refreshTokenRepository, userRepository, appRepository, log, refreshTokenTTL, accessTokenTTL)
	grpcApp := grpcapp.New(log, grpcPort, authService, rtsService)
	return &App{
		GRPCSrv: grpcApp,
	}
}
