package repository

import (
	"database/sql"
	"fmt"
)

type Transaction interface {
	Begin() error
	Rollback() error
	Commit() error
	AccounPostgrestTx() *accountRepositoryPostgres
	RefreshTokenPostgresTx() *refreshTokenRepositoryPostgres
	AccountDevicePostgresTx() *accountDeviceRepositoryPostgres
}

type sqlTransaction struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSqlTransaction(db *sql.DB) *sqlTransaction {
	return &sqlTransaction{
		db: db,
	}
}

func (s *sqlTransaction) Begin() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("[transaction][Begin][db.Begin] Error: %w", err)
	}

	s.tx = tx

	return nil
}

func (s *sqlTransaction) Rollback() error {
	return s.tx.Rollback()
}

func (s *sqlTransaction) Commit() error {
	return s.tx.Commit()
}

func (s *sqlTransaction) AccounPostgrestTx() *accountRepositoryPostgres {
	return &accountRepositoryPostgres{
		dbtx: s.tx,
	}
}

func (s *sqlTransaction) RefreshTokenPostgresTx() *refreshTokenRepositoryPostgres {
	return &refreshTokenRepositoryPostgres{
		dbtx: s.tx,
	}
}

func (s *sqlTransaction) AccountDevicePostgresTx() *accountDeviceRepositoryPostgres {
	return &accountDeviceRepositoryPostgres{
		dbtx: s.tx,
	}
}
