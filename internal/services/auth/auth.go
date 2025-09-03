package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/jwt"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/finaptica/sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository UserRepository
	appRepository  AppRepository
	log            *slog.Logger
	tokenTTL       time.Duration
}

type UserRepository interface {
	CreateUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	IsUserExistByEmail(ctx context.Context, email string) (bool, error)
}

type AppRepository interface {
	GetApp(ctx context.Context, appId int) (models.App, error)
}

// NewAuthService returns a new instance of the AuthService
func NewAuthService(log *slog.Logger, userRepository UserRepository, appRepository AppRepository, ttl time.Duration) *AuthService {
	return &AuthService{
		log:            log,
		userRepository: userRepository,
		appRepository:  appRepository,
		tokenTTL:       ttl,
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appId int) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound || errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", sl.Err(err))
			return "", errs.WithKind(op, errs.Unauthenticated, err)
		}

		log.Error("failed to get user", sl.Err(err))
		return "", errs.WithKind(op, errs.Internal, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Error("invalid password", sl.Err(err))
		return "", errs.WithKind(op, errs.Unauthenticated, err)
	}

	app, err := a.appRepository.GetApp(ctx, appId)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound || errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return "", errs.WithKind(op, errs.Unauthenticated, err)
		}

		log.Error("failed to get app", sl.Err(err))
		return "", errs.WithKind(op, errs.Internal, err)
	}
	log.Info("user logged in successfully")
	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate jwt token", sl.Err(err))
		return "", errs.WithKind(op, errs.Internal, err)
	}

	return token, nil
}

func (a *AuthService) Register(ctx context.Context, email, password string) (userId int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, errs.WithKind(op, errs.Internal, err)
	}

	id, err := a.userRepository.CreateUser(ctx, email, passHash)
	if err != nil {
		if errs.KindOf(err) == errs.AlreadyExists || errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists on create", sl.Err(err))
			return 0, errs.WithKind(op, errs.AlreadyExists, err)
		}
		log.Error("failed to create new user", sl.Err(err))
		return 0, errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user registered")
	return id, nil

}
