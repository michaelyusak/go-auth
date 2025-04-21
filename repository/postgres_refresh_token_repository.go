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

func (r *refreshTokenRepositoryPostgres) InsertToken(ctx context.Context, token string, accountId, expiredAt int64) error {
	q := `
		INSERT INTO refresh_tokens (refresh_token, account_id, expired_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
	`

	_, err := r.dbtx.ExecContext(ctx, q,
		token,
		accountId,
		expiredAt,
		nowUnixMilli())
	if err != nil {
		return err
	}

	return nil
}
