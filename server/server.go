package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelyusak/go-auth/config"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

func Init() {
	log := hHelper.NewLogrus()

	config := config.Init(log)

	router := createRouter(log, &config)

	srv := http.Server{
		Handler: router,
		Addr:    config.Port,
	}

	go func() {
		log.Infof("Sever running on port %s", config.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 10)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Infof("Server shutting down in %s ...", time.Duration(config.GracefulPeriod).String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GracefulPeriod))
	defer cancel()

	<-ctx.Done()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shut down: %s", err.Error())
	}

	log.Info("Server shut down")
}
