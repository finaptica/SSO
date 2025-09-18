package contracts

import "time"

type TokensInfo struct {
	AccessToken           string
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}
