package documenttext

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/documenttext"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(dbContext *database.DbContext) *Repository {
	return &Repository{DbContext: dbContext}
}

// ListPendingIndexing returns a keyset page (file id greater than afterFileID)
// of active document files that have no document_text row yet or whose row is
// older than the file's updated_at.
func (repository *Repository) ListPendingIndexing(afterFileID int, limit int) ([]PendingDocument, error) {
	pendingDocuments := []PendingDocument{}
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.SelectPendingIndexingQuery, pq.Array(utils.DocumentTextFormats), afterFileID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pendingDocument PendingDocument
			if err := rows.Scan(&pendingDocument.FileID, &pendingDocument.Path, &pendingDocument.Format, &pendingDocument.Size, &pendingDocument.UpdatedAt); err != nil {
				return err
			}
			pendingDocuments = append(pendingDocuments, pendingDocument)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("documenttext: list pending indexing: %w", err)
	}
	return pendingDocuments, nil
}

func (repository *Repository) UpsertDocumentText(documentText DocumentTextModel) error {
	var storedError any
	if documentText.ErrorCode != "" {
		storedError = documentText.ErrorCode
	}

	err := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			queries.UpsertDocumentTextQuery,
			documentText.FileID,
			documentText.ExtractedText,
			documentText.TextLength,
			documentText.Truncated,
			documentText.SourceUpdatedAt,
			storedError,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("documenttext: upsert document text: %w", err)
	}
	return nil
}

// SearchDocuments returns indexed documents whose text contains every term,
// name matches first, then most recently updated.
func (repository *Repository) SearchDocuments(termPatterns utils.SearchTermPatterns, limit int, offset int) ([]DocumentMatchModel, error) {
	matches := []DocumentMatchModel{}
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.SearchDocumentsQuery, termPatterns.DrivingPattern, pq.Array(termPatterns.AllPatterns), limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var match DocumentMatchModel
			if err := rows.Scan(&match.FileID, &match.Name, &match.Path, &match.ParentPath, &match.Format, &match.Size, &match.UpdatedAt, &match.ExtractedText); err != nil {
				return err
			}
			matches = append(matches, match)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("documenttext: search documents: %w", err)
	}
	return matches, nil
}
