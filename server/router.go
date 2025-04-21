package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/michaelyusak/go-auth/adaptor"
	"github.com/michaelyusak/go-auth/config"
	"github.com/michaelyusak/go-auth/handler"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-auth/service"
	helperHandler "github.com/michaelyusak/go-helper/handler"
	hHelper "github.com/michaelyusak/go-helper/helper"
	helperMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

type routerOpts struct {
	common  *helperHandler.CommonHandler
	account *handler.AccountHandler
}

func createRouter(log *logrus.Logger, config *config.ServiceConfig) *gin.Engine {
	db := adaptor.ConnectPostgres(config.Postgres, log)

	transaction := repository.NewSqlTransaction(db)
	accountRepo := repository.NewAccountRepositoryPostgres(db)
	refreshTokenRepo := repository.NewRefreshTokenRepositoryPostgres(db)
	accountDeviceRepo := repository.NewAccountDeviceRepositoryPostgres(db)

	hashHelper := hHelper.NewHashHelper(config.Hash)
	jwtHelper := hHelper.NewJWTHelper(config.Jwt.Secret)

	accountService := service.NewAccountService(service.AccountServiceOpt{
		AccountRepo:       accountRepo,
		RefreshTokenRepo:  refreshTokenRepo,
		AccountDeviceRepo: accountDeviceRepo,
		Transaction:       transaction,
		Hash:              hashHelper,
		Jwt:               jwtHelper,
		Log:               log,
	})

	commonHandler := &helperHandler.CommonHandler{}
	accountHandler := handler.NewAccountHandler(time.Duration(config.ContextTimeout), accountService)

	return newRouter(
		routerOpts{
			common:  commonHandler,
			account: accountHandler,
		},
		log,
		config.AllowedOrigins,
	)
}

func newRouter(r routerOpts, log *logrus.Logger, allowedOrigins []string) *gin.Engine {
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

	corsRouting(router, corsConfig, allowedOrigins)
	commonRouting(router, r.common)
	accountRouting(router, r.account)

	return router
}

func corsRouting(router *gin.Engine, configCors cors.Config, allowedOrigins []string) {
	configCors.AllowOrigins = allowedOrigins
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
	api.POST("/login", handler.Login)
}
