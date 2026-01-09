package service

import (
	"context"

	"github.com/michaelyusak/go-auth/entity"
	hEntity "github.com/michaelyusak/go-helper/entity"
)

type Account interface {
	Register(ctx context.Context, newAccount entity.Account) error
	Login(ctx context.Context, req entity.LoginReq) (*entity.TokenData, error)
}

type Authentication interface {
	ValidateAccessToken(ctx context.Context, accessToken string) (*hEntity.JwtCustomClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenData, error)
}
