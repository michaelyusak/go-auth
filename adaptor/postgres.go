package adaptor

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/michaelyusak/go-helper/adaptor"
	"github.com/michaelyusak/go-helper/entity"
)

func ConnectPostgres(config entity.DBConfig) (*sql.DB, error) {
	return adaptor.ConnectDB(adaptor.PSQL, config)
}
