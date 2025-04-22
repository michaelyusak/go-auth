package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/constant"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/service"
	"github.com/michaelyusak/go-helper/apperror"
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

	userAgent := ctx.Request.Header.Get(constant.UserAgentHeaderKey)
	if userAgent == "" {
		err := apperror.BadRequestError(apperror.AppErrorOpt{
			ResponseMessage: "User-Agent must not empty",
		})

		ctx.Error(err)
		return
	}

	deviceInfo := ctx.Request.Header.Get(constant.DeviceInfoHeaderKey)
	if deviceInfo == "" {
		err := apperror.BadRequestError(apperror.AppErrorOpt{
			ResponseMessage: "Device-Info must not empty",
		})

		ctx.Error(err)
		return
	}

	var newAccount entity.Account

	err := ctx.ShouldBindJSON(&newAccount)
	if err != nil {
		ctx.Error(err)
		return
	}

	c := helper.InjectValues(ctx.Request.Context(), map[any]any{
		constant.UserAgentCtxKey: userAgent,
		constant.DeviceInfoCtxKey: deviceInfo,
	})

	ctxWithTimeout, cancel := context.WithTimeout(c, h.timeout)
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

	userAgent := ctx.Request.Header.Get(constant.UserAgentHeaderKey)
	if userAgent == "" {
		err := apperror.BadRequestError(apperror.AppErrorOpt{
			ResponseMessage: "User-Agent must not empty",
		})

		ctx.Error(err)
		return
	}

	deviceInfo := ctx.Request.Header.Get(constant.DeviceInfoHeaderKey)
	if deviceInfo == "" {
		err := apperror.BadRequestError(apperror.AppErrorOpt{
			ResponseMessage: "Device-Info must not empty",
		})

		ctx.Error(err)
		return
	}

	var req entity.LoginReq

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	c := helper.InjectValues(ctx.Request.Context(), map[any]any{
		constant.UserAgentCtxKey: userAgent,
		constant.DeviceInfoCtxKey: deviceInfo,
	})

	ctxWithTimeout, cancel := context.WithTimeout(c, h.timeout)
	defer cancel()

	data, err := h.accountService.Login(ctxWithTimeout, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, data)
}
