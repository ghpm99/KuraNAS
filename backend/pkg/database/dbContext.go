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
	// Inicia a transação
	tx, err := context.database.Begin()
	if err != nil {
		return err
	}
	// Garante que o Rollback será chamado em caso de falha.
	defer tx.Rollback()

	// Executa a função do usuário dentro da transação
	if err := fn(tx); err != nil {
		return err
	}

	// Faz o commit da transação
	return tx.Commit()
}

// QueryTx executa uma função de transação de leitura com um read lock.
// A função 'fn' recebe a transação e deve retornar um erro caso falhe.
// Esta abordagem permite múltiplas leituras em paralelo.
func (context *DbContext) QueryTx(fn func(*sql.Tx) error) error {

	// Inicia a transação de leitura
	tx, err := context.database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback para transações de leitura é mais rápido que Commit

	// Executa a função do usuário dentro da transação
	if err := fn(tx); err != nil {
		return err
	}

	// Não é necessário um commit em transações de leitura
	return nil
}
