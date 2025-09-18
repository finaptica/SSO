package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/finaptica/sso/internal/config"
	"github.com/finaptica/sso/internal/handlers"
	"github.com/finaptica/sso/internal/lib/middlewares"
	"github.com/finaptica/sso/internal/services"
	"github.com/finaptica/sso/internal/storage"
	"github.com/finaptica/sso/internal/storage/repository"
	"github.com/go-chi/chi/v5"
)

type App struct {
	log    *slog.Logger
	router *chi.Mux
	port   int
}

func New(log *slog.Logger, cfg *config.Config) *App {
	db, err := storage.New(log, cfg.ConnectionStringPostgres)
	if err != nil {
		log.Error("failed to init storage", slog.String("err", err.Error()))
		panic(err)
	}
	repositoryContainer := services.RepositoriesContainer{
		UserRepo: repository.NewUserRepository(log, db),
		RtsRepo:  repository.NewRefreshTokenRepository(log, db),
		AppRepo:  repository.NewAppRepository(log, db),
		Uow:      storage.NewUnitOfWork(db),
	}

	servicesContainer := handlers.ServicesContainer{
		AuthService: services.NewAuthService(log, repositoryContainer, cfg),
		RtsService:  services.NewRefreshTokenService(log, repositoryContainer, cfg),
	}

	authHandler := handlers.NewAuthHandler(servicesContainer)

	r := chi.NewRouter()
	r.Use(middlewares.Recoverer(log))
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	return &App{
		router: r,
		port:   cfg.Http.Port,
		log:    log,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	err := http.ListenAndServe(fmt.Sprintf(":%d", a.port), a.router)
	if err != nil {
		log.Error("failed to listen and serve")
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() error {
	return nil
}
