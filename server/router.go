package server

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/config"
	"github.com/michaelyusak/go-auth/handler"
	"github.com/michaelyusak/go-auth/helper"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-auth/service"
	helperHandler "github.com/michaelyusak/go-helper/handler"
	helperMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

type routerOpts struct {
	common  *helperHandler.CommonHandler
	account *handler.AccountHandler
}

type helperOpts struct {
}

func createRouter(log *logrus.Logger, db *sql.DB, hashConfig config.HashConfig) *gin.Engine {
	transaction := repository.NewSqlTransaction(db)
	accountRepo := repository.NewAccountRepositoryPostgres(db)

	hashHelper := helper.NewHashHelperImpl(hashConfig)

	accountService := service.NewAccountService(service.AccountServiceOpt{
		AccountRepo: accountRepo,
		Transaction: transaction,
		Hash:        hashHelper,
	})

	commonHandler := &helperHandler.CommonHandler{}
	accountHandler := handler.NewAccountHandler(accountService)

	return newRouter(
		routerOpts{
			common:  commonHandler,
			account: accountHandler,
		},
		helperOpts{},
		log,
	)
}

func newRouter(r routerOpts, h helperOpts, log *logrus.Logger) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()

	router.ContextWithFallback = true

	router.Use(
		helperMiddleware.Logger(log),
		helperMiddleware.RequestIdHandlerMiddleware,
		helperMiddleware.ErrorHandlerMiddleware,
		gin.Recovery(),
	)

	// authMiddleware := middleware.AuthMiddleware(h.jwtHelper)

	corsRouting(router, corsConfig)
	commonRouting(router, r.common)

	return router
}

func corsRouting(router *gin.Engine, configCors cors.Config) {
	configCors.AllowAllOrigins = true
	configCors.AllowMethods = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}
	configCors.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Accept", "User-Agent", "Cache-Control"}
	configCors.ExposeHeaders = []string{"Content-Length"}
	configCors.AllowCredentials = true
	router.Use(cors.New(configCors))
}

func commonRouting(router *gin.Engine, handler *helperHandler.CommonHandler) {
	router.GET("/ping", handler.Ping)
	router.NoRoute(handler.NoRoute)
}

func accountRouting(router *gin.Engine, handler *handler.AccountHandler) {
	api := router.Group("v1/account")

	api.POST("/register", handler.Register)
}