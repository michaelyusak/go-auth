package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/service"
	"github.com/michaelyusak/go-helper/appconstant"
	"github.com/michaelyusak/go-helper/apperror"
	"github.com/michaelyusak/go-helper/helper"
)

type Authentication struct {
	authenticationService service.Authentication
	timeout               time.Duration
}

func NewAuthentication(authenticationService service.Authentication, timeout time.Duration) *Authentication {
	return &Authentication{
		authenticationService: authenticationService,
		timeout:               timeout,
	}
}

func (h *Authentication) ValidateAccessToken(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	accessToken := ctx.Value(appconstant.AccessTokenKey).(string)
	if accessToken == "" {
		ctx.Error(apperror.BadRequestError(apperror.AppErrorOpt{}))
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), h.timeout)
	defer cancel()

	data, err := h.authenticationService.ValidateAccessToken(ctxWithTimeout, accessToken)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, data)
}

func (h *Authentication) RefreshToken(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.RefreshTokenReq

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), h.timeout)
	defer cancel()

	data, err := h.authenticationService.RefreshToken(ctxWithTimeout, req.RefreshToken)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, data)
}
