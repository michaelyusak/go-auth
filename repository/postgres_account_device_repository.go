package repository

import (
	"context"

	"github.com/michaelyusak/go-auth/entity"
)

type accountDeviceRepositoryPostgres struct {
	dbtx DBTX
}

func NewAccountDeviceRepositoryPostgres(dbtx DBTX) *accountDeviceRepositoryPostgres {
	return &accountDeviceRepositoryPostgres{
		dbtx: dbtx,
	}
}

func (r *accountDeviceRepositoryPostgres) InsertDevice(ctx context.Context, newDevice entity.AccountDevice) error {
	q := `
		INSERT INTO account_devices (account_id, device_hash, user_agent, device_info, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`

	_, err := r.dbtx.ExecContext(ctx, q,
		newDevice.AccountId,
		newDevice.DeviceHash,
		newDevice.UserAgent,
		newDevice.DeviceInfo,
		nowUnixMilli())
	if err != nil {
		return err
	}	

	return nil
}