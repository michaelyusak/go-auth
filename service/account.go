package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/michaelyusak/go-auth/constant"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-helper/appconstant"
	"github.com/michaelyusak/go-helper/apperror"
	hEntity "github.com/michaelyusak/go-helper/entity"
	hHelper "github.com/michaelyusak/go-helper/helper"
	"github.com/sirupsen/logrus"
)

type account struct {
	accountRepo       repository.Accounts
	refreshTokenRepo  repository.RefreshTokens
	accountDeviceRepo repository.AccountDevices
	transaction       repository.Transaction
	hash              hHelper.HashHelper
	jwt               hHelper.JWTHelper
	subRoutineTimeout time.Duration
}

type AccountOpt struct {
	AccountRepo       repository.Accounts
	RefreshTokenRepo  repository.RefreshTokens
	AccountDeviceRepo repository.AccountDevices
	Transaction       repository.Transaction
	Hash              hHelper.HashHelper
	Jwt               hHelper.JWTHelper
	SubRoutineTimeout time.Duration
}

func NewAccount(opt AccountOpt) *account {
	return &account{
		accountRepo:       opt.AccountRepo,
		refreshTokenRepo:  opt.RefreshTokenRepo,
		accountDeviceRepo: opt.AccountDeviceRepo,
		transaction:       opt.Transaction,
		hash:              opt.Hash,
		jwt:               opt.Jwt,
		subRoutineTimeout: opt.SubRoutineTimeout,
	}
}

func (s *account) Register(ctx context.Context, newAccount entity.Account) error {
	if !helper.ValidatePassword(newAccount.Password) {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         constant.MsgInvalidPassword,
			ResponseMessage: constant.MsgInvalidPassword,
		})
	}

	tx, err := s.transaction.Begin()
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][transaction.Begin] Error: %s", err.Error()),
		})
	}

	accountTx := s.accountRepo.NewTx(tx)

	defer func() {
		if err != nil {
			s.transaction.Rollback()
		}

		s.transaction.Commit()
	}()

	err = accountTx.Lock(ctx)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][transaction.AccounPostgrestTx] Error: %s", err.Error()),
		})
	}

	existing, err := accountTx.GetAccountByEmail(ctx, newAccount.Email)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.GetAccountByEmail] Error: %s", err.Error()),
		})
	}
	if existing != nil {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Register] email already registered",
			ResponseMessage: "email already registered",
		})
	}

	existing, err = accountTx.GetAccountByPhoneNumber(ctx, newAccount.PhoneNumber)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.GetAccountByPhoneNumber] Error: %s", err.Error()),
		})
	}
	if existing != nil {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Register] phone number already registered",
			ResponseMessage: "phone number already registered",
		})
	}

	existing, err = accountTx.GetAccountByName(ctx, newAccount.Name)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.GetAccountByName] Error: %s", err.Error()),
		})
	}
	if existing != nil {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Register] name already registered",
			ResponseMessage: "name already registered",
		})
	}

	hash, err := s.hash.Hash(newAccount.Password)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][hash.Hash] passwordHash | Error: %s", err.Error()),
		})
	}

	newAccount.Password = hash

	accountId, err := accountTx.Register(ctx, newAccount)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.Register] Error: %s | account_id: %v", err.Error(), accountId),
		})
	}

	newAccount.Id = accountId

	userAgent := ctx.Value(appconstant.UserAgentKey).(string)
	deviceInfo := ctx.Value(appconstant.DeviceInfokey).(string)

	deviceHash := hHelper.GenerateDeviceHash(ctx, accountId)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.subRoutineTimeout)
		defer cancel()

		accountDevice := entity.AccountDevice{
			AccountId:  newAccount.Id,
			DeviceHash: deviceHash,
			UserAgent:  userAgent,
			DeviceInfo: deviceInfo,
		}

		_, err = s.accountDeviceRepo.InsertDevice(ctx, accountDevice)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":       err.Error(),
				"account_id":  newAccount.Id,
				"device_hash": deviceHash,
			}).Error("[account_service][Register][accountDeviceRepo.InsertDevice][sub-routine]")
		}
	}()

	return nil
}

func (s *account) Login(ctx context.Context, req entity.LoginReq) (*entity.TokenData, error) {
	if req.Email == "" && req.Name == "" {
		return nil, apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Login] either email or name must be provided",
			ResponseMessage: "either email or name must be provided",
		})
	}

	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][transaction.Begin] Error: %s", err.Error()),
		})
	}

	accountTx := s.accountRepo.NewTx(tx)
	refreshTokenTx := s.refreshTokenRepo.NewTx(tx)
	accountDeviceTx := s.accountDeviceRepo.NewTx(tx)

	defer func() {
		if err != nil {
			s.transaction.Rollback()
		}

		s.transaction.Commit()
	}()

	err = accountTx.Lock(ctx)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][transaction.AccounPostgrestTx] Error: %s", err.Error()),
		})
	}

	var account *entity.Account

	if req.Email != "" {
		account, err = accountTx.GetAccountByEmail(ctx, req.Email)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountTx.GetAccountByEmail] Error: %s | email: %s", err.Error(), req.Email),
			})
		}
	} else if req.Name != "" {
		account, err = accountTx.GetAccountByName(ctx, req.Name)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountTx.GetAccountByName] Error: %s | name: %s", err.Error(), req.Name),
			})
		}
	}

	if account == nil {
		return nil, apperror.NewAppError(apperror.AppErrorOpt{
			Code:            http.StatusForbidden,
			Message:         fmt.Sprintf("[account_service][Login] account not found | email: %s | name: %s", req.Email, req.Name),
			ResponseMessage: constant.MsgAccountNotFound,
		})
	}

	isValid, err := s.hash.Check(req.Password, []byte(account.Password))
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][hash.Check] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	if !isValid {
		return nil, apperror.UnauthorizedError(apperror.AppErrorOpt{
			Message:         fmt.Sprintf("[account_service][Login] invalid credentials | account_id: %v", account.Id),
			ResponseMessage: constant.MsgInvalidLogin,
		})
	}

	accountDeviceHash := hHelper.GenerateDeviceHash(ctx, account.Id)

	accountDevice, err := accountDeviceTx.GetDeviceByHashAndAccountId(ctx, accountDeviceHash, account.Id)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][accountDeviceTx.GetDeviceByHash] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	if accountDevice == nil {
		err = refreshTokenTx.DeleteTokenByAccountId(ctx, account.Id)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][refreshTokenTx.DeleteTokenByAccountId] Error: %s | account_id: %v", err.Error(), account.Id),
			})
		}

		userAgent := ctx.Value(appconstant.UserAgentKey).(string)
		deviceInfo := ctx.Value(appconstant.DeviceInfokey).(string)

		newDevice := entity.AccountDevice{
			AccountId:  account.Id,
			DeviceHash: accountDeviceHash,
			UserAgent:  userAgent,
			DeviceInfo: deviceInfo,
		}

		newDeviceId, err := accountDeviceTx.InsertDevice(ctx, newDevice)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountDeviceTx.InsertDevice] Error: %s | account_id: %v", err.Error(), account.Id),
			})
		}

		accountDevice = &newDevice
		accountDevice.DeviceId = newDeviceId
	}

	customClaims := hEntity.JwtCustomClaims{
		AccountId: account.Id,
		Email:     account.Email,
		Name:      account.Name,
		DeviceId:  accountDevice.DeviceId,
	}

	customClaimsBytes, err := json.Marshal(customClaims)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][json.Marshal] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	accessTokenExpiredAt := time.Now().Add(30 * time.Minute).Unix()

	accessToken, err := s.jwt.CreateAndSign(customClaimsBytes, accessTokenExpiredAt)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][jwt.CreateAndSign][Access] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	refreshTokenExpiredAt := time.Now().Add(24 * time.Hour).Unix()

	refreshToken, err := s.jwt.CreateAndSign(customClaimsBytes, refreshTokenExpiredAt)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][jwt.CreateAndSign][Refresh] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.subRoutineTimeout)
		defer cancel()

		err := s.refreshTokenRepo.InsertToken(ctx, refreshToken, account.Id, accountDevice.DeviceId, refreshTokenExpiredAt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"account_id": account.Id,
			}).Error("[account_service][Login][refreshTokenRepo.InsertToken][sub-routine]")
		}
	}()

	return &entity.TokenData{
		AccessToken: &entity.Token{
			Token:     accessToken,
			ExpiredAt: accessTokenExpiredAt,
		},
		RefreshToken: &entity.Token{
			Token:     refreshToken,
			ExpiredAt: refreshTokenExpiredAt,
		},
	}, nil
}
