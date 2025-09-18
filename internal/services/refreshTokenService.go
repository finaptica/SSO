package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/config"
	"github.com/finaptica/sso/internal/contracts"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	tokenGen "github.com/finaptica/sso/internal/lib/token"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenService struct {
	refreshTokenRepository RefreshTokenRepository
	userRepository         UserRepository
	appRepository          AppRepository
	uow                    UnitOfWork
	log                    *slog.Logger
	refreshTokenTTL        time.Duration
	accessTokenTTl         time.Duration
}

// NewRefreshTokenService() returns a new instance of a RefreshTokenService
func NewRefreshTokenService(log *slog.Logger, repoCont RepositoriesContainer, cfg *config.Config) *RefreshTokenService {
	return &RefreshTokenService{
		refreshTokenRepository: repoCont.RtsRepo,
		userRepository:         repoCont.UserRepo,
		appRepository:          repoCont.AppRepo,
		uow:                    repoCont.Uow,
		log:                    log,
		refreshTokenTTL:        cfg.RefreshTokenTTL,
		accessTokenTTl:         cfg.AccessTokenTTL,
	}
}

func (rts *RefreshTokenService) RefreshTokens(ctx context.Context, refreshToken string) (contracts.TokensInfo, error) {
	const op = "refreshTokenService.RefreshTokens"
	log := rts.log.With(slog.String("op", op))

	var result contracts.TokensInfo

	err := rts.uow.Do(ctx, func(tx pgx.Tx) error {
		token, err := rts.refreshTokenRepository.GetByValueTx(ctx, tx, refreshToken)
		if err != nil {
			return errs.WithKind(op, errs.NotFound, err)
		}

		if token.IsRevoked || token.ExpiresAt.Before(time.Now().UTC()) {
			return errs.WithKind(op, errs.Unauthenticated, errors.New("invalid token"))
		}

		if err := rts.refreshTokenRepository.RevokeTx(ctx, tx, token.ID); err != nil {
			return errs.WithKind(op, errs.Internal, err)
		}

		newValue := tokenGen.NewRefreshToken()
		newExp := time.Now().UTC().Add(rts.refreshTokenTTL)
		if _, err := rts.refreshTokenRepository.SaveNewRefreshTokenTx(ctx, tx, token.UserID, token.AppID, newValue, newExp); err != nil {
			return errs.WithKind(op, errs.Internal, err)
		}

		user, err := rts.userRepository.GetUserByIDTx(ctx, tx, token.UserID)
		if err != nil {
			return errs.WithKind(op, errs.Internal, err)
		}

		app, err := rts.appRepository.GetAppByIDTx(ctx, tx, token.AppID)
		if err != nil {
			return errs.WithKind(op, errs.Internal, err)
		}

		accessToken, err := tokenGen.NewAccessToken(user, app, rts.accessTokenTTl)
		if err != nil {
			return errs.WithKind(op, errs.Internal, err)
		}

		result = contracts.TokensInfo{
			AccessToken:           accessToken,
			RefreshToken:          newValue,
			RefreshTokenExpiresAt: newExp,
		}
		return nil
	})

	if err != nil {
		log.Error("failed to refresh tokens", sl.Err(err))
		return contracts.TokensInfo{}, err
	}

	return result, nil
}
