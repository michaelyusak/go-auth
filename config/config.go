package config

import (
	"fmt"
	"os"

	hConfig "github.com/michaelyusak/go-helper/config"
	hEntity "github.com/michaelyusak/go-helper/entity"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

type JwtConfig struct {
	Secret               hHelper.JwtConfig `json:"secret"`
	AccessTokenDuration  hEntity.Duration  `json:"access_token_duration"`
	RefreshTokenDuration hEntity.Duration  `json:"refresh_token_duration"`
}

type ContextTimeoutConfig struct {
	Main       hEntity.Duration `json:"main"`
	SubRoutine hEntity.Duration `json:"sub_routine"`
}

type AuthConfig struct {
	AllowedIpAddress  []string `json:"allowed_ip_address"`
	AllowedDeviceInfo []string `json:"allowed_device_info"`
}

type AppConfig struct {
	Port           string               `json:"port"`
	LogLevel       string               `json:"log_level"`
	GracefulPeriod hEntity.Duration     `json:"graceful_period"`
	ContextTimeout ContextTimeoutConfig `json:"context_timeout"`
	Postgres       hEntity.DBConfig     `json:"postgres"`
	Jwt            JwtConfig            `json:"jwt"`
	Hash           hHelper.HashConfig   `json:"hash"`
	AllowedOrigins []string             `json:"allowed_origins"`
	Auth           AuthConfig           `json:"auth"`
}

func Init() (AppConfig, error) {
	configPath := os.Getenv("GO_AUTH_SERVICE_CONFIG")

	var conf AppConfig

	conf, err := hConfig.InitFromJson[AppConfig](configPath)
	if err != nil {
		return conf, fmt.Errorf("[config][Init][hConfig.InitFromJson] Failed to init config from json: %w", err)
	}

	return conf, nil
}
