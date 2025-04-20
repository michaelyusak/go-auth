package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/service"
	"github.com/michaelyusak/go-helper/helper"
)

type AccountHandler struct {
	timeout        time.Duration
	accountService service.AccountService
}

func NewAccountHandler(timeout time.Duration, accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		timeout:        timeout,
		accountService: accountService,
	}
}

func (h *AccountHandler) Register(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var newAccount entity.Account

	err := ctx.ShouldBindJSON(&newAccount)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), h.timeout)
	defer cancel()

	err = h.accountService.Register(ctxWithTimeout, newAccount)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, nil)
}

func (h *AccountHandler) Login(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.LoginReq

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), h.timeout)
	defer cancel()

	err = h.accountService.Login(ctxWithTimeout, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, nil)
}
