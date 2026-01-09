package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelyusak/go-auth/adaptor"
	"github.com/michaelyusak/go-auth/config"
	"github.com/michaelyusak/go-auth/handler"
	"github.com/michaelyusak/go-auth/repository"
	"github.com/michaelyusak/go-auth/repository/postgres"
	"github.com/michaelyusak/go-auth/service"
	hHandler "github.com/michaelyusak/go-helper/handler"
	hHelper "github.com/michaelyusak/go-helper/helper"
	hMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

var (
	APP_HEALTHY = false
)

type routerOpts struct {
	common  *hHandler.Common
	account *handler.Account
	auth    *handler.Authentication
}

func createRouter(config *config.AppConfig) *gin.Engine {
	db, err := adaptor.ConnectPostgres(config.Postgres)
	if err != nil {
		logrus.Panicf("Failed to connect to db: %v", err)
	}
	logrus.Info("Connected to postgres")

	transaction := repository.NewSqlTransaction(db)
	accountsRepo := postgres.NewAccounts(db)
	refreshTokensRepo := postgres.NewRefreshTokens(db)
	accountDevicesRepo := postgres.NewAccountDevices(db)

	hashHelper := hHelper.NewHashHelper(config.Hash)
	jwtHelper := hHelper.NewJWTHelper(config.Jwt.Secret, jwt.SigningMethodHS512)

	accountService := service.NewAccount(service.AccountOpt{
		AccountRepo:       accountsRepo,
		RefreshTokenRepo:  refreshTokensRepo,
		AccountDeviceRepo: accountDevicesRepo,
		Transaction:       transaction,
		Hash:              hashHelper,
		Jwt:               jwtHelper,
		SubRoutineTimeout: time.Duration(config.ContextTimeout.SubRoutine),
	})
	authenticationService := service.NewAuthentication(service.AuthenticationOpt{
		AccountDevicesRepo: accountDevicesRepo,
		AccountsRepo:       accountsRepo,
		Transaction:        transaction,
		Jwt:                jwtHelper,
	})

	commonHandler := hHandler.NewCommon(&APP_HEALTHY)
	accountHandler := handler.NewAccount(time.Duration(config.ContextTimeout.Main), accountService)
	authenticationHandler := handler.NewAuthentication(authenticationService, time.Duration(config.ContextTimeout.Main))

	return newRouter(
		routerOpts{
			common:  commonHandler,
			account: accountHandler,
			auth:    authenticationHandler,
		},
		config.AllowedOrigins,
		config.Auth,
	)
}

func newRouter(r routerOpts, allowedOrigins []string, authConfig config.AuthConfig) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()

	router.ContextWithFallback = true

	router.Use(
		hMiddleware.Logger(logrus.New()),
		hMiddleware.RequestIdHandlerMiddleware,
		hMiddleware.ErrorHandlerMiddleware,
		gin.Recovery(),
	)

	authMiddleware := hMiddleware.NewAuth(hMiddleware.AuthOpt{
		IsCheckDeviceId:   true,
		AllowedIpAddress:  authConfig.AllowedIpAddress,
		AllowedDeviceInfo: authConfig.AllowedDeviceInfo,
	}).Auth()

	corsRouting(router, corsConfig, allowedOrigins)
	commonRouting(router, r.common)
	accountRouting(router, r.account, authMiddleware)
	authenticationRouting(router, r.auth, authMiddleware)

	return router
}

func corsRouting(router *gin.Engine, configCors cors.Config, allowedOrigins []string) {
	configCors.AllowOrigins = allowedOrigins
	configCors.AllowMethods = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}
	configCors.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Accept", "User-Agent", "Cache-Control", "Device-Info", "X-Device-Id"}
	configCors.ExposeHeaders = []string{"Content-Length"}
	configCors.AllowCredentials = true
	router.Use(cors.New(configCors))
}

func commonRouting(router *gin.Engine, handler *hHandler.Common) {
	router.GET("/health", handler.Health)
	router.NoRoute(handler.NoRoute)
}

func accountRouting(router *gin.Engine, handler *handler.Account, authMiddleware gin.HandlerFunc) {
	api := router.Group("v1/account")

	api.POST("/register", authMiddleware, handler.Register)
	api.POST("/login", authMiddleware, handler.Login)
}

func authenticationRouting(router *gin.Engine, handler *handler.Authentication, authMiddleware gin.HandlerFunc) {
	api := router.Group("v1/auth")

	api.POST("/validate-token", authMiddleware, handler.ValidateAccessToken)
	api.POST("/refresh-token", authMiddleware, handler.RefreshToken)
}
