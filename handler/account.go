package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/service"
	"github.com/michaelyusak/go-helper/helper"
)

type Account struct {
	timeout        time.Duration
	accountService service.Account
}

func NewAccount(timeout time.Duration, accountService service.Account) *Account {
	return &Account{
		timeout:        timeout,
		accountService: accountService,
	}
}

func (h *Account) Register(ctx *gin.Context) {
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

func (h *Account) Login(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.LoginReq

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), h.timeout)
	defer cancel()

	data, err := h.accountService.Login(ctxWithTimeout, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, data)
}
