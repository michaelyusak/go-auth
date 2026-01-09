package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelyusak/go-auth/config"
	"github.com/michaelyusak/go-auth/log"
	"github.com/sirupsen/logrus"
)

func Init() {
	config, err := config.Init()
	if err != nil {
		logrus.Panic(err)
	}

	log.Init(config.LogLevel)

	router := createRouter(&config)

	srv := http.Server{
		Handler: router,
		Addr:    config.Port,
	}

	go func() {
		logrus.Infof("Sever running on port %s", config.Port)
		APP_HEALTHY = true

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Error while listening")
		}
	}()

	quit := make(chan os.Signal, 10)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logrus.Infof("Server shutting down in %s ...", time.Duration(config.GracefulPeriod).String())

	APP_HEALTHY = false

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GracefulPeriod))
	defer cancel()

	<-ctx.Done()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server shut down")
	}

	logrus.Info("Server shut down")
}
