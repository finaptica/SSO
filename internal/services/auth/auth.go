package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/jwt"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/finaptica/sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	log          *slog.Logger
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userId int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appId int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// NewAuthService returns a new instance of the AuthService
func NewAuthService(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, ttl time.Duration) *AuthService {
	return &AuthService{
		log:          log,
		userProvider: userProvider,
		appProvider:  appProvider,
		userSaver:    userSaver,
		tokenTTL:     ttl,
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appId int) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		a.log.Error("invalid passowrd", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get app", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	log.Info("user loggen in successfully")
	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate jwt token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *AuthService) RegisterNewUser(ctx context.Context, email, password string) (userId int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save new user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")
	return id, nil

}

func (a *AuthService) IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	isAdmin, err = a.userProvider.IsAdmin(ctx, userId)
	if err != nil {
		log.Info("")
	}
}
