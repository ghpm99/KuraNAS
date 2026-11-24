package documenttext

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/documenttext"

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
		rows, err := tx.Query(queries.SelectPendingIndexingQuery, pq.Array(IndexableFormats()), afterFileID, limit)
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

	var storedFailureReason any
	if documentText.FailureReason != "" {
		storedFailureReason = documentText.FailureReason
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
			documentText.State,
			documentText.IndexedBytes,
			storedFailureReason,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("documenttext: upsert document text: %w", err)
	}
	return nil
}

// FindDocumentForIndexing returns the active file when it has a document
// format (07/RN-01) and whether it exists.
func (repository *Repository) FindDocumentForIndexing(fileID int) (PendingDocument, bool, error) {
	var document PendingDocument
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.SelectDocumentForIndexingQuery, fileID, pq.Array(DocumentFormats())).
			Scan(&document.FileID, &document.Path, &document.Format, &document.Size, &document.UpdatedAt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return PendingDocument{}, false, nil
	}
	if err != nil {
		return PendingDocument{}, false, fmt.Errorf("documenttext: find document: %w", err)
	}
	return document, true, nil
}

// FindIndexState returns the indexing state of an active document; a document
// without a row, or with a row older than the file, is queued.
func (repository *Repository) FindIndexState(fileID int) (IndexStateModel, bool, error) {
	var indexState IndexStateModel
	var failureReason sql.NullString
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.SelectIndexStateQuery, fileID, pq.Array(DocumentFormats())).
			Scan(&indexState.State, &indexState.IndexedBytes, &failureReason)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return IndexStateModel{}, false, nil
	}
	if err != nil {
		return IndexStateModel{}, false, fmt.Errorf("documenttext: find index state: %w", err)
	}
	indexState.FailureReason = failureReason.String
	return indexState, true, nil
}
