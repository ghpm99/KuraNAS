package email

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/email"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{DbContext: db}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

// scanAccount reads the listing shape (no token column).
func scanAccount(scan func(dest ...any) error) (AccountModel, error) {
	var m AccountModel
	var provider, status string
	var lastSyncAt sql.NullTime

	if err := scan(&m.ID, &provider, &m.Address, &m.DisplayName, &status, &m.SyncEnabled, &lastSyncAt, &m.LastError, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return AccountModel{}, err
	}

	m.Provider = Provider(provider)
	m.Status = AccountStatus(status)
	if lastSyncAt.Valid {
		m.LastSyncAt = &lastSyncAt.Time
	}
	return m, nil
}

func (r *Repository) ListAccounts() ([]AccountModel, error) {
	var accounts []AccountModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListAccountsQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			m, scanErr := scanAccount(rows.Scan)
			if scanErr != nil {
				return scanErr
			}
			accounts = append(accounts, m)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("ListAccounts: %w", err)
	}

	return accounts, nil
}

func (r *Repository) GetAccountByID(id int) (AccountModel, error) {
	var m AccountModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var provider, status string
		var lastSyncAt sql.NullTime

		scanErr := tx.QueryRow(queries.GetAccountByIDQuery, id).Scan(
			&m.ID, &provider, &m.Address, &m.DisplayName, &m.TokenCiphertext,
			&status, &m.SyncEnabled, &lastSyncAt, &m.LastError, &m.CreatedAt, &m.UpdatedAt,
		)
		if scanErr != nil {
			return scanErr
		}

		m.Provider = Provider(provider)
		m.Status = AccountStatus(status)
		if lastSyncAt.Valid {
			m.LastSyncAt = &lastSyncAt.Time
		}
		return nil
	})

	if err != nil {
		return AccountModel{}, fmt.Errorf("GetAccountByID: %w", err)
	}

	return m, nil
}

func (r *Repository) UpsertAccount(model AccountModel) (int, error) {
	var id int

	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			queries.UpsertAccountQuery,
			string(model.Provider),
			model.Address,
			model.DisplayName,
			model.TokenCiphertext,
		).Scan(&id)
	})

	if err != nil {
		return 0, fmt.Errorf("UpsertAccount: %w", err)
	}

	return id, nil
}

func (r *Repository) UpdateAccountTokens(id int, tokenCiphertext []byte, status AccountStatus, lastError string) error {
	err := r.execExpectingRow(queries.UpdateAccountTokensQuery, id, tokenCiphertext, string(status), lastError)
	if err != nil {
		return fmt.Errorf("UpdateAccountTokens: %w", err)
	}
	return nil
}

func (r *Repository) UpdateSyncEnabled(id int, enabled bool) error {
	err := r.execExpectingRow(queries.UpdateAccountSyncEnabledQuery, id, enabled)
	if err != nil {
		return fmt.Errorf("UpdateSyncEnabled: %w", err)
	}
	return nil
}

func (r *Repository) DeleteAccount(id int) error {
	err := r.execExpectingRow(queries.DeleteAccountQuery, id)
	if err != nil {
		return fmt.Errorf("DeleteAccount: %w", err)
	}
	return nil
}

// execExpectingRow runs a statement that must touch exactly one row and maps
// "no rows touched" to sql.ErrNoRows so the service can answer 404.
func (r *Repository) execExpectingRow(query string, args ...any) error {
	return r.DbContext.ExecTx(func(tx *sql.Tx) error {
		result, err := tx.Exec(query, args...)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}
