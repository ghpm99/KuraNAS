package database

import (
	"database/sql"
)

type DbContext struct {
	database *sql.DB
}

func NewDbContext(db *sql.DB) *DbContext {
	return &DbContext{
		database: db,
	}
}

func (context *DbContext) GetDatabase() *sql.DB {
	return context.database
}

func (context *DbContext) ExecTx(fn func(*sql.Tx) error) error {
	tx, err := context.database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (context *DbContext) QueryTx(fn func(*sql.Tx) error) error {

	tx, err := context.database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return nil
}
