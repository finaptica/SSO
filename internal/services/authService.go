package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/config"
	"github.com/finaptica/sso/internal/contracts"
	"github.com/finaptica/sso/internal/lib/errs"
	tokenGen "github.com/finaptica/sso/internal/lib/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository         UserRepository
	appRepository          AppRepository
	refreshTokenRepository RefreshTokenRepository
	log                    *slog.Logger
	accessTokenTTL         time.Duration
	refreshTokenTTL        time.Duration
}

// NewAuthService returns a new instance of the AuthService
func NewAuthService(log *slog.Logger, repoContainer RepositoriesContainer, cfg *config.Config) *AuthService {
	return &AuthService{
		log:                    log,
		userRepository:         repoContainer.UserRepo,
		appRepository:          repoContainer.AppRepo,
		refreshTokenRepository: repoContainer.RtsRepo,
		accessTokenTTL:         cfg.AccessTokenTTL,
		refreshTokenTTL:        cfg.RefreshTokenTTL,
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appId int) (tokensInfo contracts.TokensInfo, err error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound {
			return contracts.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
		}

		return contracts.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		return contracts.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
	}

	app, err := a.appRepository.GetAppById(ctx, appId)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound {
			return contracts.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
		}

		return contracts.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	accessToken, err := tokenGen.NewAccessToken(user, app, a.accessTokenTTL)
	if err != nil {
		return contracts.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	refreshTokenValue := tokenGen.NewRefreshToken()
	refreshTokenExpiresAt := time.Now().UTC().Add(a.refreshTokenTTL)
	_, err = a.refreshTokenRepository.SaveNewRefreshToken(ctx, user.ID, app.ID, refreshTokenValue, refreshTokenExpiresAt)
	if err != nil {
		return contracts.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user logged in successfully")
	tokensInfo = contracts.TokensInfo{
		AccessToken:           accessToken,
		RefreshToken:          refreshTokenValue,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}

	return tokensInfo, nil
}

func (a *AuthService) Register(ctx context.Context, email, password string) (userId uuid.UUID, err error) {
	const op = "auth.Register"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, errs.WithKind(op, errs.Internal, err)
	}

	id, err := a.userRepository.CreateUser(ctx, email, passHash)
	if err != nil {
		if errs.KindOf(err) == errs.AlreadyExists {
			return uuid.UUID{}, errs.WithKind(op, errs.AlreadyExists, err)
		}
		return uuid.UUID{}, errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user registered")
	return id, nil

}
