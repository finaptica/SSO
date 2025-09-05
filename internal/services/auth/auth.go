package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	srv "github.com/finaptica/sso/internal/grpc/auth"
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

type UserRepository interface {
	CreateUser(ctx context.Context, email string, passHash []byte) (uid uuid.UUID, err error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type AppRepository interface {
	GetApp(ctx context.Context, appId int) (models.App, error)
}

type RefreshTokenRepository interface {
	SaveNewRefreshToken(ctx context.Context, userId uuid.UUID, appId int, value string, expiresAt time.Time) error
}

// NewAuthService returns a new instance of the AuthService
func NewAuthService(log *slog.Logger, userRepository UserRepository, appRepository AppRepository, refreshTokenRepository RefreshTokenRepository, acessTTL time.Duration, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		log:                    log,
		userRepository:         userRepository,
		appRepository:          appRepository,
		refreshTokenRepository: refreshTokenRepository,
		accessTokenTTL:         acessTTL,
		refreshTokenTTL:        refreshTTL,
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appId int) (tokensInfo srv.TokensInfo, err error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound {
			return srv.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
		}

		return srv.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		return srv.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
	}

	app, err := a.appRepository.GetApp(ctx, appId)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound {
			return srv.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, err)
		}

		return srv.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	accessToken, err := tokenGen.NewAccessToken(user, app, a.accessTokenTTL)
	if err != nil {
		return srv.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	refreshTokenValue := tokenGen.NewRefreshToken()
	refreshTokenExpiresAt := time.Now().UTC().Add(a.refreshTokenTTL)
	err = a.refreshTokenRepository.SaveNewRefreshToken(ctx, user.ID, app.ID, refreshTokenValue, refreshTokenExpiresAt)
	if err != nil {
		return srv.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user logged in successfully")
	tokensInfo = srv.TokensInfo{
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
