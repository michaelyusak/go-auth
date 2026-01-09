package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
)

type accountDevices struct {
	dbtx repository.DBTX
}

func NewAccountDevices(dbtx repository.DBTX) *accountDevices {
	return &accountDevices{
		dbtx: dbtx,
	}
}

func (r *accountDevices) NewTx(tx *sql.Tx) repository.AccountDevices {
	return &accountDevices{
		dbtx: tx,
	}
}

func (r *accountDevices) InsertDevice(ctx context.Context, newDevice entity.AccountDevice) (int64, error) {
	q := `
		INSERT INTO account_devices (account_id, device_hash, user_agent, device_info, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING device_id
	`

	var deviceId int64

	err := r.dbtx.QueryRowContext(ctx, q,
		newDevice.AccountId,
		newDevice.DeviceHash,
		newDevice.UserAgent,
		newDevice.DeviceInfo,
		helper.NowUnixMilli()).Scan(&deviceId)
	if err != nil {
		return deviceId, err
	}

	return deviceId, nil
}

func (r *accountDevices) GetDeviceByHashAndAccountId(ctx context.Context, hash string, accountId int64) (*entity.AccountDevice, error) {
	q := `
		SELECT device_id, account_id, device_hash, user_agent, device_info, created_at, updated_at, deleted_at
		FROM account_devices
		WHERE device_hash = $1
			AND account_id = $2
			AND deleted_at IS NULL
	`

	var accountDevice entity.AccountDevice

	err := r.dbtx.QueryRowContext(ctx, q, hash, accountId).Scan(
		&accountDevice.DeviceId,
		&accountDevice.AccountId,
		&accountDevice.DeviceHash,
		&accountDevice.UserAgent,
		&accountDevice.DeviceInfo,
		&accountDevice.CreatedAt,
		&accountDevice.UpdatedAt,
		&accountDevice.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &accountDevice, nil
}
