package repository

import (
	"context"
)

type refreshTokenRepositoryPostgres struct {
	dbtx DBTX
}

func NewRefreshTokenRepositoryPostgres(dbtx DBTX) *refreshTokenRepositoryPostgres {
	return &refreshTokenRepositoryPostgres{
		dbtx: dbtx,
	}
}

func (r *refreshTokenRepositoryPostgres) InsertToken(ctx context.Context, token string, accountId, deviceId, expiredAt int64) error {
	q := `
		INSERT INTO refresh_tokens (refresh_token, account_id, device_id, expired_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`

	_, err := r.dbtx.ExecContext(ctx, q,
		token,
		accountId,
		deviceId,
		expiredAt,
		nowUnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (r *refreshTokenRepositoryPostgres) DeleteTokenByAccountId(ctx context.Context, accountId int64) error {
	q := `
		DELETE FROM refresh_tokens
		WHERE account_id = $1
	`

	_, err := r.dbtx.ExecContext(ctx, q, accountId)
	if err != nil {
		return err
	}

	return nil
}
