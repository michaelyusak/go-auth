package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-helper/apperror"
	hEntity "github.com/michaelyusak/go-helper/entity"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

type authentication struct {
	accountDevicesRepo repository.AccountDevices
	accountsRepo       repository.Accounts

	transaction repository.Transaction
	jwt         hHelper.JWTHelper
}

type AuthenticationOpt struct {
	AccountDevicesRepo repository.AccountDevices
	AccountsRepo       repository.Accounts

	Transaction repository.Transaction
	Jwt         hHelper.JWTHelper
}

func NewAuthentication(opt AuthenticationOpt) *authentication {
	return &authentication{
		accountDevicesRepo: opt.AccountDevicesRepo,
		accountsRepo:       opt.AccountsRepo,

		transaction: opt.Transaction,
		jwt:         opt.Jwt,
	}
}

func (s *authentication) ValidateAccessToken(ctx context.Context, accessToken string) (*hEntity.JwtCustomClaims, error) {
	customClaimsBytes, err := s.jwt.ParseAndVerify(accessToken)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][ValidateAccessToken][jwt.ParseAndVerify] Error: %s", err.Error()),
		})
	}
	if customClaimsBytes == nil {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	var customClaims hEntity.JwtCustomClaims

	err = json.Unmarshal(customClaimsBytes, &customClaims)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][ValidateAccessToken][json.Unmarshal] Error: %s", err.Error()),
		})
	}

	deviceHash := hHelper.GenerateDeviceHash(ctx, customClaims.AccountId)

	accountDevice, err := s.accountDevicesRepo.GetDeviceByHashAndAccountId(ctx, deviceHash, customClaims.AccountId)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][ValidateAccessToken][accountDevicesRepo.GetDeviceByHash] Error: %s", err.Error()),
		})
	}

	if accountDevice == nil || accountDevice.DeviceId != customClaims.DeviceId {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	account, err := s.accountsRepo.GetAccountById(ctx, customClaims.AccountId)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][ValidateAccessToken][accountsRepo.GetAccountById] Error: %s", err.Error()),
		})
	}

	if account == nil || account.Email != customClaims.Email || account.Name != customClaims.Name {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	return &customClaims, err
}

func (s *authentication) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenData, error) {
	customClaimsBytes, err := s.jwt.ParseAndVerify(refreshToken)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][RefreshToken][jwt.ParseAndVerify] Error: %s", err.Error()),
		})
	}
	if customClaimsBytes == nil {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	var customClaims hEntity.JwtCustomClaims

	err = json.Unmarshal(customClaimsBytes, &customClaims)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][RefreshToken][json.Unmarshal] Error: %s", err.Error()),
		})
	}

	deviceHash := hHelper.GenerateDeviceHash(ctx, customClaims.AccountId)

	accountDevice, err := s.accountDevicesRepo.GetDeviceByHashAndAccountId(ctx, deviceHash, customClaims.AccountId)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][RefreshToken][accountDevicesRepo.GetDeviceByHash] Error: %s", err.Error()),
		})
	}

	if accountDevice == nil || accountDevice.DeviceId != customClaims.DeviceId {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	account, err := s.accountsRepo.GetAccountById(ctx, customClaims.AccountId)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][RefreshToken][accountsRepo.GetAccountById] Error: %s", err.Error()),
		})
	}

	if account == nil || account.Email != customClaims.Email || account.Name != customClaims.Name {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{})
	}

	accessTokenExpiredAt := time.Now().Add(30 * time.Minute).Unix()

	newAccessToken, err := s.jwt.CreateAndSign(customClaimsBytes, accessTokenExpiredAt)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[authentication_service][RefreshToken][jwt.CreateAndSign] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	return &entity.TokenData{
		AccessToken: &entity.Token{
			Token:     newAccessToken,
			ExpiredAt: accessTokenExpiredAt,
		},
	}, nil
}
