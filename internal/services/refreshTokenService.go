package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/grpc/auth"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	tokenGen "github.com/finaptica/sso/internal/lib/token"
)

type RefreshTokenService struct {
	refreshTokenRepository RefreshTokenRepository
	userRepository         UserRepository
	appRepository          AppRepository
	log                    *slog.Logger
	refreshTokenTTL        time.Duration
	accessTokenTTl         time.Duration
}

// NewRefreshTokenService() returns a new instance of a RefreshTokenService
func NewRefreshTokenService(refreshTokenRepo RefreshTokenRepository, uR UserRepository, aR AppRepository, log *slog.Logger, refreshTokenTTL time.Duration, accessTokenTTL time.Duration) *RefreshTokenService {
	return &RefreshTokenService{
		refreshTokenRepository: refreshTokenRepo,
		userRepository:         uR,
		appRepository:          aR,
		log:                    log,
		refreshTokenTTL:        refreshTokenTTL,
		accessTokenTTl:         accessTokenTTL,
	}
}

func (rts *RefreshTokenService) RefreshTokens(ctx context.Context, refreshToken string) (auth.TokensInfo, error) {
	const op = "refreshTokenService.RefreshTokens"
	log := rts.log.With(slog.String("op", op))

	tx, err := rts.refreshTokenRepository.DB().BeginTxx(ctx, nil)
	if err != nil {
		log.Error("failed to begin transaction", sl.Err(err))
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	token, err := rts.refreshTokenRepository.GetByValueTx(ctx, tx, refreshToken)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.NotFound, err)
	}

	if token.IsRevoked || token.ExpiresAt.Before(time.Now().UTC()) {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Unauthenticated, errors.New("invalid token"))
	}

	err = rts.refreshTokenRepository.RevokeTx(ctx, tx, token.ID)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	newValue := tokenGen.NewRefreshToken()
	newExp := time.Now().UTC().Add(rts.refreshTokenTTL)
	err = rts.refreshTokenRepository.SaveNewRefreshTokenTx(ctx, tx, token.UserID, token.AppID, newValue, newExp)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	user, err := rts.userRepository.GetUserByIDTx(ctx, tx, token.UserID)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	app, err := rts.appRepository.GetAppByIDTx(ctx, tx, token.AppID)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	accessToken, err := tokenGen.NewAccessToken(user, app, rts.accessTokenTTl)
	if err != nil {
		return auth.TokensInfo{}, errs.WithKind(op, errs.Internal, err)
	}

	return auth.TokensInfo{
		AccessToken:           accessToken,
		RefreshToken:          newValue,
		RefreshTokenExpiresAt: newExp,
	}, nil
}
