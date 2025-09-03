package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/jwt"
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
		if errs.KindOf(err) == errs.NotFound {
			return "", errs.WithKind(op, errs.Unauthenticated, err)
		}

		return "", errs.WithKind(op, errs.Internal, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		return "", errs.WithKind(op, errs.Unauthenticated, err)
	}

	app, err := a.appRepository.GetApp(ctx, appId)
	if err != nil {
		if errs.KindOf(err) == errs.NotFound {
			return "", errs.WithKind(op, errs.Unauthenticated, err)
		}

		return "", errs.WithKind(op, errs.Internal, err)
	}

	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		return "", errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user logged in successfully")

	return token, nil
}

func (a *AuthService) Register(ctx context.Context, email, password string) (userId int64, err error) {
	const op = "auth.Register"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errs.WithKind(op, errs.Internal, err)
	}

	id, err := a.userRepository.CreateUser(ctx, email, passHash)
	if err != nil {
		if errs.KindOf(err) == errs.AlreadyExists {
			return 0, errs.WithKind(op, errs.AlreadyExists, err)
		}
		return 0, errs.WithKind(op, errs.Internal, err)
	}

	log.Info("user registered")
	return id, nil

}
