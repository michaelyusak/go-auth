package service

import (
	"context"
	"fmt"

	"github.com/michaelyusak/go-auth/constant"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-helper/apperror"
)

type accountServiceImpl struct {
	accountRepo repository.AccountRepository
	transaction repository.Transaction
	hash        helper.HashHelper
}

type AccountServiceOpt struct {
	AccountRepo repository.AccountRepository
	Transaction repository.Transaction
	Hash        helper.HashHelper
}

func NewAccountService(opt AccountServiceOpt) *accountServiceImpl {
	return &accountServiceImpl{
		accountRepo: opt.AccountRepo,
		transaction: opt.Transaction,
		hash:        opt.Hash,
	}
}

func (s *accountServiceImpl) Register(ctx context.Context, newAccount entity.Account) error {
	if newAccount.Email == "" && newAccount.PhoneNumber == "" {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "either email or phone number must be provided",
			ResponseMessage: "either email or phone number must be provided",
		})
	}

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

	var existing *entity.Account

	if newAccount.Email != "" {
		existing, err = accountRepo.GetAccountByEmail(ctx, newAccount.Email)
		if err != nil {
			return apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Register][accountRepo.GetAccountByEmail] Error: %s", err.Error()),
			})
		}
	} else if newAccount.PhoneNumber != "" {
		existing, err = accountRepo.GetAccountByPhoneNumber(ctx, newAccount.PhoneNumber)
		if err != nil {
			return apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Register][accountRepo.GetAccountByPhoneNumber] Error: %s", err.Error()),
			})
		}
	}

	if existing != nil {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Register] email already registered",
			ResponseMessage: "email already registered",
		})
	}

	hash, err := s.hash.Hash(newAccount.Password)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][hash.Hash] Error: %s", err.Error()),
		})
	}

	newAccount.Password = hash

	err = accountRepo.Register(ctx, newAccount)
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Register][accountRepo.Register] Error: %s", err.Error()),
		})
	}

	return nil
}
