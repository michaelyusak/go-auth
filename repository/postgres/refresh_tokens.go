package postgres

import (
	"context"
	"database/sql"

	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
)

type refreshTokens struct {
	dbtx repository.DBTX
}

func NewRefreshTokens(dbtx repository.DBTX) *refreshTokens {
	return &refreshTokens{
		dbtx: dbtx,
	}
}

func (r *refreshTokens) NewTx(tx *sql.Tx) repository.RefreshTokens {
	return &refreshTokens{
		dbtx: tx,
	}
}

func (r *refreshTokens) InsertToken(ctx context.Context, token string, accountId, deviceId, expiredAt int64) error {
	q := `
		INSERT INTO refresh_tokens (refresh_token, account_id, device_id, expired_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`

	_, err := r.dbtx.ExecContext(ctx, q,
		token,
		accountId,
		deviceId,
		expiredAt,
		helper.NowUnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (r *refreshTokens) DeleteTokenByAccountId(ctx context.Context, accountId int64) error {
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
