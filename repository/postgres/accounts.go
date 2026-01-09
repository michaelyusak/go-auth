package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
)

type accounts struct {
	dbtx repository.DBTX
}

func NewAccounts(dbtx repository.DBTX) *accounts {
	return &accounts{
		dbtx: dbtx,
	}
}

func (r *accounts) NewTx(tx *sql.Tx) repository.Accounts {
	return &accounts{
		dbtx: tx,
	}
}

func (r *accounts) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	q := `
		SELECT account_id, account_name, account_email, account_phone_number, account_password, created_at, updated_at, deleted_at
		FROM accounts
		WHERE account_email = $1
			AND deleted_at IS NULL
	`

	var account entity.Account

	err := r.dbtx.QueryRowContext(ctx, q, email).Scan(
		&account.Id,
		&account.Name,
		&account.Email,
		&account.PhoneNumber,
		&account.Password,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[postgres][account_repository][GetAccountByEmail][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}

func (r *accounts) GetAccountByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.Account, error) {
	q := `
	SELECT account_id, account_name, account_email, account_phone_number, account_password, created_at, updated_at, deleted_at
	FROM accounts
	WHERE account_phone_number = $1
		AND deleted_at IS NULL
	`

	var account entity.Account

	err := r.dbtx.QueryRowContext(ctx, q, phoneNumber).Scan(
		&account.Id,
		&account.Name,
		&account.Email,
		&account.PhoneNumber,
		&account.Password,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[postgres][account_repository][GetAccountByPhoneNumber][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}

func (r *accounts) Lock(ctx context.Context) error {
	q := `
		LOCK TABLE accounts IN EXCLUSIVE MODE;
	`

	_, err := r.dbtx.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("[postgres][account_repository][Lock][ExecContext] Error: %w", err)
	}

	return nil
}

func (r *accounts) Register(ctx context.Context, newAccount entity.Account) (int64, error) {
	q := `
		INSERT INTO accounts (account_name, account_email, account_phone_number, account_password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING account_id
	`

	var accountId int64

	err := r.dbtx.QueryRowContext(ctx, q,
		newAccount.Name,
		newAccount.Email,
		newAccount.PhoneNumber,
		newAccount.Password,
		helper.NowUnixMilli()).Scan(&accountId)
	if err != nil {
		return accountId, fmt.Errorf("[postgres][account_repository][Register][QueryRowContext] Error: %w", err)
	}

	return accountId, nil
}

func (r *accounts) GetAccountByName(ctx context.Context, name string) (*entity.Account, error) {
	q := `
	SELECT account_id, account_name, account_email, account_phone_number, account_password, created_at, updated_at, deleted_at
	FROM accounts
	WHERE account_name = $1
		AND deleted_at IS NULL
	`

	var account entity.Account

	err := r.dbtx.QueryRowContext(ctx, q, name).Scan(
		&account.Id,
		&account.Name,
		&account.Email,
		&account.PhoneNumber,
		&account.Password,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[postgres][account_repository][GetAccountByName][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}

func (r *accounts) GetAccountById(ctx context.Context, accountId int64) (*entity.Account, error) {
	q := `
	SELECT account_id, account_name, account_email, account_phone_number, account_password, created_at, updated_at, deleted_at
	FROM accounts
	WHERE account_id = $1
		AND deleted_at IS NULL
	`

	var account entity.Account

	err := r.dbtx.QueryRowContext(ctx, q, accountId).Scan(
		&account.Id,
		&account.Name,
		&account.Email,
		&account.PhoneNumber,
		&account.Password,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[postgres][account_repository][GetAccountById][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}
