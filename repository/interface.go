package repository

import (
	"context"
	"database/sql"

	"github.com/michaelyusak/go-auth/entity"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type AccountRepository interface {
	GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error)
	GetAccountByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.Account, error)
	Lock(ctx context.Context) error
	Register(ctx context.Context, newAccount entity.Account) (int64, error)
	GetAccountByName(ctx context.Context, name string) (*entity.Account, error)
}

type RefreshTokenRepository interface {
	InsertToken(ctx context.Context, token string, accountId, expiredAt int64) error
}

type AccountDeviceRepository interface {
	InsertDevice(ctx context.Context, newDevice entity.AccountDevice) error
}
