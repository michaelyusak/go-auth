package server

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	helperHandler "github.com/michaelyusak/go-helper/handler"
	helperMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

type routerOpts struct {
	common *helperHandler.CommonHandler
}

type helperOpts struct {
}

func createRouter(log *logrus.Logger, db *sql.DB) *gin.Engine {
	commonHandler := &helperHandler.CommonHandler{}

	return newRouter(
		routerOpts{
			common: commonHandler,
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
