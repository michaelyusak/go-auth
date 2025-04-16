package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelyusak/go-auth/adaptor"
	"github.com/michaelyusak/go-auth/config"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

func Init() {
	log := hHelper.NewLogrus()

	config := config.Init(log)

	db := adaptor.ConnectPostgres(config.Postgres, log)

	router := createRouter(log, db)

	srv := http.Server{
		Handler: router,
		Addr:    config.Port,
	}

	go func() {
		log.Infof("Sever running on port: %s", config.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 10)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Server shutdown gracefully ...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GracefulPeriod)*time.Second)
	defer cancel()

	<-ctx.Done()

	log.Infof("Timeout of %v seconds", config.GracefulPeriod)

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown: %s", err.Error())
	}

	log.Info("Server exited")
}
