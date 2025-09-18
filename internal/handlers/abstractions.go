package handlers

import (
	"context"

	"github.com/finaptica/sso/internal/contracts"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (userId uuid.UUID, err error)
	Login(ctx context.Context, email string, password string, appId int) (tokensInfo contracts.TokensInfo, err error)
}

type RefreshTokenService interface {
	RefreshTokens(ctx context.Context, refreshToken string) (contracts.TokensInfo, error)
}

type ServicesContainer struct {
	AuthService AuthService
	RtsService  RefreshTokenService
}
