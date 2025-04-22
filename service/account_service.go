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
	"github.com/michaelyusak/go-helper/apperror"
	hHelper "github.com/michaelyusak/go-helper/helper"
	"github.com/sirupsen/logrus"
)

type accountServiceImpl struct {
	accountRepo       repository.AccountRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	accountDeviceRepo repository.AccountDeviceRepository
	transaction       repository.Transaction
	hash              hHelper.HashHelper
	jwt               hHelper.JWTHelper
	log               *logrus.Logger
	subRoutineTimeout time.Duration
}

type AccountServiceOpt struct {
	AccountRepo       repository.AccountRepository
	RefreshTokenRepo  repository.RefreshTokenRepository
	AccountDeviceRepo repository.AccountDeviceRepository
	Transaction       repository.Transaction
	Hash              hHelper.HashHelper
	Jwt               hHelper.JWTHelper
	Log               *logrus.Logger
	SubRoutineTimeout time.Duration
}

func NewAccountService(opt AccountServiceOpt) *accountServiceImpl {
	return &accountServiceImpl{
		accountRepo:       opt.AccountRepo,
		refreshTokenRepo:  opt.RefreshTokenRepo,
		accountDeviceRepo: opt.AccountDeviceRepo,
		transaction:       opt.Transaction,
		hash:              opt.Hash,
		jwt:               opt.Jwt,
		log:               opt.Log,
		subRoutineTimeout: opt.SubRoutineTimeout,
	}
}

func (s *accountServiceImpl) Register(ctx context.Context, newAccount entity.Account) error {
	if !helper.ValidatePassword(newAccount.Password) {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         constant.MsgInvalidPassword,
			ResponseMessage: constant.MsgInvalidPassword,
		})
	}

	err := s.transaction.Begin()
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][transaction.Begin] Error: %s", err.Error()),
		})
	}

	accountRepo := s.transaction.AccounPostgrestTx()

	defer func() {
		if err != nil {
			s.transaction.Rollback()
		}

		s.transaction.Commit()
	}()

	err = accountRepo.Lock(ctx)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][transaction.AccounPostgrestTx] Error: %s", err.Error()),
		})
	}

	existing, err := accountRepo.GetAccountByEmail(ctx, newAccount.Email)
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

	existing, err = accountRepo.GetAccountByPhoneNumber(ctx, newAccount.PhoneNumber)
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

	existing, err = accountRepo.GetAccountByName(ctx, newAccount.Name)
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

	accountId, err := accountRepo.Register(ctx, newAccount)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.Register] Error: %s | account_id: %v", err.Error(), accountId),
		})
	}

	newAccount.Id = accountId

	userAgent := ctx.Value(constant.UserAgentCtxKey).(string)
	deviceInfo := ctx.Value(constant.DeviceInfoCtxKey).(string)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.subRoutineTimeout)
		defer cancel()

		deviceHash := s.hash.HashSHA512(fmt.Sprintf("%v%s%s", newAccount.Id, userAgent, deviceInfo))

		accountDevice := entity.AccountDevice{
			AccountId:  newAccount.Id,
			DeviceHash: deviceHash,
			UserAgent:  userAgent,
			DeviceInfo: deviceInfo,
		}

		_, err = s.accountDeviceRepo.InsertDevice(ctx, accountDevice)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"error":       err.Error(),
				"account_id":  newAccount.Id,
				"device_hash": deviceHash,
			}).Error("[account_service][Register][accountDeviceRepo.InsertDevice][sub-routine]")
		}
	}()

	return nil
}

func (s *accountServiceImpl) Login(ctx context.Context, req entity.LoginReq) (*entity.TokenData, error) {
	if req.Email == "" && req.Name == "" {
		return nil, apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Login] either email or name must be provided",
			ResponseMessage: "either email or name must be provided",
		})
	}

	err := s.transaction.Begin()
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][transaction.Begin] Error: %s", err.Error()),
		})
	}

	accountRepo := s.transaction.AccounPostgrestTx()
	refreshTokenRepo := s.transaction.RefreshTokenPostgresTx()
	accountDeviceRepo := s.transaction.AccountDevicePostgresTx()

	defer func() {
		if err != nil {
			s.transaction.Rollback()
		}

		s.transaction.Commit()
	}()

	err = accountRepo.Lock(ctx)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][transaction.AccounPostgrestTx] Error: %s", err.Error()),
		})
	}

	var account *entity.Account

	if req.Email != "" {
		account, err = accountRepo.GetAccountByEmail(ctx, req.Email)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountRepo.GetAccountByEmail] Error: %s | email: %s", err.Error(), req.Email),
			})
		}
	} else if req.Name != "" {
		account, err = accountRepo.GetAccountByName(ctx, req.Name)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountRepo.GetAccountByName] Error: %s | name: %s", err.Error(), req.Name),
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

	userAgent := ctx.Value(constant.UserAgentCtxKey).(string)
	deviceInfo := ctx.Value(constant.DeviceInfoCtxKey).(string)

	accountDeviceHash := s.hash.HashSHA512(fmt.Sprintf("%v%s%s", account.Id, userAgent, deviceInfo))

	accountDevice, err := accountDeviceRepo.GetDeviceByHash(ctx, accountDeviceHash)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][accountDeviceRepo.GetDeviceByHash] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	if accountDevice == nil {
		err = refreshTokenRepo.DeleteTokenByAccountId(ctx, account.Id)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][refreshTokenRepo.DeleteTokenByAccountId] Error: %s | account_id: %v", err.Error(), account.Id),
			})
		}

		newDevice := entity.AccountDevice{
			AccountId:  account.Id,
			DeviceHash: accountDeviceHash,
			UserAgent:  userAgent,
			DeviceInfo: deviceInfo,
		}

		newDeviceId, err := accountDeviceRepo.InsertDevice(ctx, newDevice)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountDeviceRepo.InsertDevice] Error: %s | account_id: %v", err.Error(), account.Id),
			})
		}

		accountDevice = &newDevice
		accountDevice.DeviceId = newDeviceId
	}

	customClaims := make(map[string]any)
	customClaims["account_id"] = account.Id
	customClaims["email"] = account.Email
	customClaims["name"] = account.Name

	customClaimsBytes, err := json.Marshal(customClaims)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][json.Marshal] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	accessTokenExpiredAt := time.Now().Add(30 * time.Minute).UnixMilli()

	accessToken, err := s.jwt.CreateAndSign(customClaimsBytes, accessTokenExpiredAt)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][json.Marshal][Access] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	refreshTokenExpiredAt := time.Now().Add(24 * time.Hour).UnixMilli()

	refreshToken, err := s.jwt.CreateAndSign(customClaimsBytes, refreshTokenExpiredAt)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][json.Marshal][Refresh] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.subRoutineTimeout)
		defer cancel()

		err := s.refreshTokenRepo.InsertToken(ctx, refreshToken, account.Id, accountDevice.DeviceId, refreshTokenExpiredAt)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"account_id": account.Id,
			}).Error("[account_service][Login][refreshTokenRepo.InsertToken][sub-routine]")
		}
	}()

	return &entity.TokenData{
		AccessToken: entity.Token{
			Token:     accessToken,
			ExpiredAt: accessTokenExpiredAt,
		},
		RefreshToken: entity.Token{
			Token:     refreshToken,
			ExpiredAt: refreshTokenExpiredAt,
		},
	}, nil
}
