package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/michaelyusak/go-auth/entity"
	hHelper "github.com/michaelyusak/go-helper/helper"
	"github.com/sirupsen/logrus"
)

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DbName   string `json:"db_name"`
}

type JwtConfig struct {
	Secret               hHelper.JwtConfig `json:"secret"`
	AccessTokenDuration  entity.Duration   `json:"access_token_duration"`
	RefreshTokenDuration entity.Duration   `json:"refresh_token_duration"`
}

type ServiceConfig struct {
	Port                     string             `json:"port"`
	GracefulPeriod           entity.Duration    `json:"graceful_period"`
	ContextTimeout           entity.Duration    `json:"context_timeout"`
	SubRoutineContextTimeout entity.Duration    `json:"sub_routine_context_timeout"`
	Postgres                 DBConfig           `json:"postgres"`
	Jwt                      JwtConfig          `json:"jwt"`
	Hash                     hHelper.HashConfig `json:"hash"`
	AllowedOrigins           []string           `json:"allowed_origins"`
}

func Init(log *logrus.Logger) ServiceConfig {
	configPath := os.Getenv("GO_AUTH_SERVICE_CONFIG")

	var config ServiceConfig

	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[config][Init][os.ReadFile] error: %s", err.Error()),
		}).Fatal("error initiating config file")

		return config
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[config][Init][json.Unmarshal] error: %s", err.Error()),
		}).Fatal("error initiating config file")

		return config
	}

	return config
}
