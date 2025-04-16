package repository

import (
	"context"
	"fmt"

	"github.com/michaelyusak/go-auth/entity"
)

type accountRepositoryPostgres struct {
	dbtx DBTX
}

func NewAccountRepositoryPostgres(dbtx DBTX) *accountRepositoryPostgres {
	return &accountRepositoryPostgres{
		dbtx: dbtx,
	}
}

func (r *accountRepositoryPostgres) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
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
		return nil, fmt.Errorf("[postgres][account_repository][GetAccountByEmail][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}

func (r *accountRepositoryPostgres) GetAccountByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.Account, error) {
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
		return nil, fmt.Errorf("[postgres][account_repository][GetAccountByPhoneNumber][QueryRowContext] Error: %w", err)
	}

	return &account, nil
}

func (r *accountRepositoryPostgres) Lock(ctx context.Context) error {
	q := `
		LOCK TABLE accounts IN EXCLUSIVE MODE;
	`

	_, err := r.dbtx.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("[postgres][account_repository][Lock][ExecContext] Error: %w", err)
	}

	return nil
}

func (r *accountRepositoryPostgres) Register(ctx context.Context, newAccount entity.Account) error {
	q := `
		INSERT INTO accounts (account_email, account_phone_number, account_password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
	`

	_, err := r.dbtx.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("[postgres][account_repository][Register][ExecContext] Error: %w", err)
	}

	return nil
}
