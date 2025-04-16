package adaptor

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/michaelyusak/go-auth/config"
	"github.com/sirupsen/logrus"
)

func ConnectPostgres(config config.DBConfig, log *logrus.Logger) *sql.DB {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.Username, config.Password, config.Host, config.Port, config.DbName))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[adaptor][ConnectPostgres][Open] error: %s", err.Error()),
		}).Fatal("error connecting to postgres")

		return nil
	}

	if err = db.Ping(); err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("[adaptor][ConnectPostgres][Ping] error: %s", err.Error()),
		}).Fatal("error connecting to postgres")

		return nil
	}

	return db
}
