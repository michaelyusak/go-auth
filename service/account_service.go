package service

import (
	"context"
	"fmt"
	"net/http"

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

func (s *accountServiceImpl) Login(ctx context.Context, req entity.LoginReq) error {
	if req.Email == "" && req.Name == "" {
		return apperror.BadRequestError(apperror.AppErrorOpt{
			Message:         "[account_service][Login] either email or name must be provided",
			ResponseMessage: "either email or name must be provided",
		})
	}

	err := s.transaction.Begin()
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][transaction.Begin] Error: %s", err.Error()),
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
			Message: fmt.Sprintf("[account_service][Login][transaction.AccounPostgrestTx] Error: %s", err.Error()),
		})
	}

	var account *entity.Account

	if req.Email != "" {
		account, err = accountRepo.GetAccountByEmail(ctx, req.Email)
		if err != nil {
			return apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountRepo.GetAccountByEmail] Error: %s | email: %s", err.Error(), req.Email),
			})
		}
	} else if req.Name != "" {
		account, err = accountRepo.GetAccountByName(ctx, req.Name)
		if err != nil {
			return apperror.InternalServerError(apperror.AppErrorOpt{
				Message: fmt.Sprintf("[account_service][Login][accountRepo.GetAccountByName] Error: %s | name: %s", err.Error(), req.Name),
			})
		}
	}

	if account == nil {
		return apperror.NewAppError(apperror.AppErrorOpt{
			Code:            http.StatusForbidden,
			Message:         fmt.Sprintf("[account_service][Login] account not found | email: %s | name: %s", req.Email, req.Name),
			ResponseMessage: constant.MsgAccountNotFound,
		})
	}

	isValid, err := s.hash.Check(req.Password, []byte(account.Password))
	if err != nil {
		return apperror.InternalServerError(apperror.AppErrorOpt{
			Message: fmt.Sprintf("[account_service][Login][hash.Check] Error: %s | account_id: %v", err.Error(), account.Id),
		})
	}

	if !isValid {
		return apperror.UnauthorizedError(apperror.AppErrorOpt{
			Message:         fmt.Sprintf("[account_service][Login] invalid credentials | account_id: %v", account.Id),
			ResponseMessage: constant.MsgInvalidLogin,
		})
	}

	return nil
}
