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

type Transaction interface {
	Begin() (*sql.Tx, error)
	Rollback() error
	Commit() error
}

type Accounts interface {
	NewTx(tx *sql.Tx) Accounts
	GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error)
	GetAccountByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.Account, error)
	Lock(ctx context.Context) error
	Register(ctx context.Context, newAccount entity.Account) (int64, error)
	GetAccountByName(ctx context.Context, name string) (*entity.Account, error)
	GetAccountById(ctx context.Context, accountId int64) (*entity.Account, error)
}

type RefreshTokens interface {
	NewTx(tx *sql.Tx) RefreshTokens
	InsertToken(ctx context.Context, token string, accountId, deviceId, expiredAt int64) error
	DeleteTokenByAccountId(ctx context.Context, accountId int64) error
}

type AccountDevices interface {
	NewTx(tx *sql.Tx) AccountDevices
	InsertDevice(ctx context.Context, newDevice entity.AccountDevice) (int64, error)
	GetDeviceByHashAndAccountId(ctx context.Context, hash string, accountId int64) (*entity.AccountDevice, error)
}
