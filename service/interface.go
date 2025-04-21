package service

import (
	"context"

	"github.com/michaelyusak/go-auth/entity"
)

type AccountService interface {
	Register(ctx context.Context, newAccount entity.Account) error
	Login(ctx context.Context, req entity.LoginReq) (*entity.TokenData, error)
}
