package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/entity"
	"github.com/michaelyusak/go-auth/service"
	"github.com/michaelyusak/go-helper/helper"
)

type AccountHandler struct {
	accountService service.AccountService
}

func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
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

	err = h.accountService.Register(ctx.Request.Context(), newAccount)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, nil)

}
