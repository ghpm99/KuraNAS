package email

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nas-go/api/pkg/database"
	confqueries "nas-go/api/pkg/database/queries/configuration"
	queries "nas-go/api/pkg/database/queries/email"
	"nas-go/api/pkg/utils"
)

// providerSettingKey is the app_settings key holding which AI provider analyzes
// e-mail (task 16). The e-mail domain reuses the shared settings table through
// the existing get/upsert queries instead of owning a parallel table.
const providerSettingKey = "email_ai_provider"

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

// UpdateAccountLastSync advances the per-account sync cursor after a successful
// fetch (which only happens with a valid token, so the account is linked).
func (r *Repository) UpdateAccountLastSync(id int, syncedAt time.Time) error {
	err := r.execExpectingRow(queries.UpdateAccountLastSyncQuery, id, syncedAt)
	if err != nil {
		return fmt.Errorf("UpdateAccountLastSync: %w", err)
	}
	return nil
}

// InsertMessage stores one synced message. It reports inserted=false (without
// error) when the message already exists, so the sync stays idempotent.
func (r *Repository) InsertMessage(message MessageModel) (inserted bool, err error) {
	authJSON, err := json.Marshal(message.AuthResults)
	if err != nil {
		return false, fmt.Errorf("InsertMessage: marshal auth_results: %w", err)
	}

	err = r.DbContext.ExecTx(func(tx *sql.Tx) error {
		var id int
		scanErr := tx.QueryRow(
			queries.InsertMessageQuery,
			message.AccountID,
			message.ProviderMessageID,
			message.SenderName,
			message.SenderAddress,
			message.Subject,
			message.Snippet,
			message.SanitizedBody,
			message.ReceivedAt,
			authJSON,
			marshalJSONArray(message.Attachments),
			marshalJSONArray(message.LinkDomains),
			marshalJSONArray(message.PrefilterRules),
			string(messageStatusOrDefault(message.Status)),
		).Scan(&id)
		if scanErr == sql.ErrNoRows {
			return nil // conflict: already stored
		}
		if scanErr != nil {
			return scanErr
		}
		inserted = true
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("InsertMessage: %w", err)
	}
	return inserted, nil
}

// ListMessages returns one lean page (no body) ordered newest-first. It fetches
// pageSize+1 rows so the shared pagination helper can decide HasNext.
func (r *Repository) ListMessages(page, pageSize int) (utils.PaginationResponse[MessageModel], error) {
	response := utils.PaginationResponse[MessageModel]{
		Items:      []MessageModel{},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListMessagesQuery, pageSize+1, utils.CalculateOffset(page, pageSize))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m MessageModel
			var status, verdict, importance string
			if scanErr := rows.Scan(
				&m.ID, &m.AccountID, &m.SenderName, &m.SenderAddress, &m.Subject,
				&m.Snippet, &m.ReceivedAt, &status, &m.CreatedAt,
				&verdict, &importance, &m.Summary,
			); scanErr != nil {
				return scanErr
			}
			m.Status = MessageStatus(status)
			m.Verdict = Verdict(verdict)
			m.Importance = Importance(importance)
			response.Items = append(response.Items, m)
		}
		return rows.Err()
	})

	if err != nil {
		return response, fmt.Errorf("ListMessages: %w", err)
	}

	response.UpdatePagination()
	return response, nil
}

// ListPendingMessages returns messages awaiting the pre-filter, carrying only
// the fields the deterministic rules evaluate.
func (r *Repository) ListPendingMessages(limit int) ([]MessageModel, error) {
	var messages []MessageModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListPendingMessagesQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m MessageModel
			var authJSON, attachmentJSON, linkJSON []byte
			if scanErr := rows.Scan(
				&m.ID, &m.SenderAddress, &m.Subject, &authJSON, &attachmentJSON, &linkJSON,
			); scanErr != nil {
				return scanErr
			}
			if err := unmarshalJSONIfPresent(authJSON, &m.AuthResults); err != nil {
				return err
			}
			if err := unmarshalJSONIfPresent(attachmentJSON, &m.Attachments); err != nil {
				return err
			}
			if err := unmarshalJSONIfPresent(linkJSON, &m.LinkDomains); err != nil {
				return err
			}
			m.Status = MsgStatusPending
			messages = append(messages, m)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("ListPendingMessages: %w", err)
	}
	return messages, nil
}

func (r *Repository) UpdateMessagePrefilter(id int, status MessageStatus, rules []string) error {
	err := r.execExpectingRow(queries.UpdateMessagePrefilterQuery, id, string(status), marshalJSONArray(rules))
	if err != nil {
		return fmt.Errorf("UpdateMessagePrefilter: %w", err)
	}
	return nil
}

// PurgeMessagesBefore deletes messages older than the cutoff and returns how
// many rows were removed.
func (r *Repository) PurgeMessagesBefore(cutoff time.Time) (int, error) {
	var removed int64
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		result, err := tx.Exec(queries.PurgeMessagesBeforeQuery, cutoff)
		if err != nil {
			return err
		}
		removed, err = result.RowsAffected()
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("PurgeMessagesBefore: %w", err)
	}
	return int(removed), nil
}

// ListMessagesForAnalysis returns pending messages with the full sanitized body
// and every evidence column, for the AI analysis step.
func (r *Repository) ListMessagesForAnalysis(limit int) ([]MessageModel, error) {
	var messages []MessageModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListMessagesForAnalysisQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m MessageModel
			var authJSON, attachmentJSON, linkJSON, rulesJSON []byte
			if scanErr := rows.Scan(
				&m.ID, &m.AccountID, &m.SenderName, &m.SenderAddress, &m.Subject,
				&m.SanitizedBody, &authJSON, &attachmentJSON, &linkJSON, &rulesJSON,
			); scanErr != nil {
				return scanErr
			}
			if err := unmarshalJSONIfPresent(authJSON, &m.AuthResults); err != nil {
				return err
			}
			if err := unmarshalJSONIfPresent(attachmentJSON, &m.Attachments); err != nil {
				return err
			}
			if err := unmarshalJSONIfPresent(linkJSON, &m.LinkDomains); err != nil {
				return err
			}
			if err := unmarshalJSONIfPresent(rulesJSON, &m.PrefilterRules); err != nil {
				return err
			}
			m.Status = MsgStatusPending
			messages = append(messages, m)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("ListMessagesForAnalysis: %w", err)
	}
	return messages, nil
}

// UpsertAnalysis persists one message's verdict, replacing any previous one.
func (r *Repository) UpsertAnalysis(model AnalysisModel) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(
			queries.UpsertAnalysisQuery,
			model.MessageID,
			string(model.Verdict),
			model.RiskScore,
			marshalJSONArray(model.Evidence),
			model.Summary,
			string(model.Importance),
			model.ProviderUsed,
			model.ModelUsed,
		)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("UpsertAnalysis: %w", err)
	}
	return nil
}

// UpdateMessageAnalyzed advances a message's status and expunges its sanitized
// body (retention rule A7).
func (r *Repository) UpdateMessageAnalyzed(id int, status MessageStatus) error {
	err := r.execExpectingRow(queries.UpdateMessageAnalyzedQuery, id, string(status))
	if err != nil {
		return fmt.Errorf("UpdateMessageAnalyzed: %w", err)
	}
	return nil
}

// GetAnalysisByMessage returns the stored verdict for one message, or
// sql.ErrNoRows when it has not been analyzed.
func (r *Repository) GetAnalysisByMessage(messageID int) (AnalysisModel, error) {
	var m AnalysisModel
	m.MessageID = messageID

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var verdict, importance string
		var evidenceJSON []byte
		scanErr := tx.QueryRow(queries.GetAnalysisByMessageQuery, messageID).Scan(
			&m.MessageID, &verdict, &m.RiskScore, &evidenceJSON, &m.Summary,
			&importance, &m.ProviderUsed, &m.ModelUsed,
		)
		if scanErr != nil {
			return scanErr
		}
		m.Verdict = Verdict(verdict)
		m.Importance = Importance(importance)
		return unmarshalJSONIfPresent(evidenceJSON, &m.Evidence)
	})

	if err != nil {
		return AnalysisModel{}, fmt.Errorf("GetAnalysisByMessage: %w", err)
	}
	return m, nil
}

// GetProviderPreference reads which AI provider analyzes e-mail. A missing row
// means the operator never chose one, reported as empty so the service applies
// its default (local Ollama).
func (r *Repository) GetProviderPreference() (string, error) {
	var value string
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var key, stored string
		scanErr := tx.QueryRow(confqueries.GetSettingQuery, providerSettingKey).Scan(&key, &stored)
		if scanErr != nil {
			return scanErr
		}
		value = stored
		return nil
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("GetProviderPreference: %w", err)
	}
	return value, nil
}

// SetProviderPreference stores which AI provider analyzes e-mail.
func (r *Repository) SetProviderPreference(value string) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(confqueries.UpsertSettingQuery, providerSettingKey, value)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("SetProviderPreference: %w", err)
	}
	return nil
}

// marshalJSONArray marshals a slice to JSON, emitting "[]" for nil/empty so the
// JSONB columns never hold SQL NULL or the literal "null".
func marshalJSONArray[T any](items []T) []byte {
	if len(items) == 0 {
		return []byte("[]")
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return []byte("[]")
	}
	return encoded
}

func unmarshalJSONIfPresent(raw []byte, out any) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode jsonb column: %w", err)
	}
	return nil
}

func messageStatusOrDefault(status MessageStatus) MessageStatus {
	if status == "" {
		return MsgStatusPending
	}
	return status
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
