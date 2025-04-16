package config

import (
	"encoding/json"
	"fmt"
	"os"

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
	Issuer string `json:"issuer"`
	Key    string `json:"key"`
}

type HashConfig struct {
	HashCost int `json:"hash_cost"`
}

type ServiceConfig struct {
	Port           string     `json:"port"`
	GracefulPeriod int64      `json:"graceful_perion_s"`
	ContextTimeout int64      `json:"context_timeout_s"`
	Postgres       DBConfig   `json:"postgres"`
	Jwt            JwtConfig  `json:"jwt"`
	Hash           HashConfig `json:"hash"`
}

func Init(log *logrus.Logger) ServiceConfig {
	configPath := os.Getenv("GO_AUTH_SERVICE_CONFIG")

	var config ServiceConfig

	configFile, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[config][Init][OpenFile] error: %s", err.Error()),
		}).Fatal("error initiating config file")

		return config
	}
	defer configFile.Close()

	var configData []byte

	_, err = configFile.Read(configData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[config][Init][Read] error: %s", err.Error()),
		}).Fatal("error initiating config file")

		return config
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[config][Init][Unmarshal] error: %s", err.Error()),
		}).Fatal("error initiating config file")

		return config
	}

	return config
}
