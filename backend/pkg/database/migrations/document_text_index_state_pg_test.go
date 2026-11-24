package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestDocumentTextIndexStateMigrationBackfillsExistingRows_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	type legacyRow struct {
		name         string
		text         string
		truncated    bool
		errorCode    any
		wantState    string
		wantBytes    int64
		wantReason   string
		wantRequeued bool
	}
	legacyRows := []legacyRow{
		{name: "indexed.txt", text: "olá mundo", wantState: "indexed", wantBytes: 10},
		{name: "partial.txt", text: "abcdef", truncated: true, wantState: "partial", wantBytes: 6, wantRequeued: true},
		{name: "scan.pdf", text: "", wantState: "name_only", wantReason: "no_text"},
		{name: "locked.pdf", text: "", errorCode: "encrypted", wantState: "name_only", wantReason: "encrypted"},
		{name: "huge.pdf", text: "", errorCode: "too_large", wantState: "name_only", wantReason: "too_large"},
		{name: "old.doc", text: "", errorCode: "unsupported_format", wantState: "name_only", wantReason: "unsupported_format"},
		{name: "binary.txt", text: "", errorCode: "binary", wantState: "failed", wantReason: "binary"},
		{name: "broken.docx", text: "", errorCode: "extract_failed", wantState: "failed", wantReason: "extract_failed"},
	}

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM home_file WHERE parent_path = '/legacy_docs'`); err != nil {
			return err
		}
		for _, row := range legacyRows {
			var fileID int
			err := tx.QueryRow(
				`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at)
				 VALUES ($1::text, '/legacy_docs/' || $1::text, '/legacy_docs', '.txt', 1, '2026-01-01', '2026-01-01') RETURNING id`,
				row.name,
			).Scan(&fileID)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(
				`INSERT INTO document_text (file_id, extracted_text, text_length, truncated, source_updated_at, error, state, indexed_bytes, failure_reason)
				 VALUES ($1, $2, 0, $3, '2026-01-01', $4, 'indexed', 0, NULL)`,
				fileID, row.text, row.truncated, row.errorCode,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed legacy rows: %v", seedErr)
	}
	t.Cleanup(func() {
		_ = dbContext.ExecTx(func(tx *sql.Tx) error {
			_, err := tx.Exec(`DELETE FROM home_file WHERE parent_path = '/legacy_docs'`)
			return err
		})
	})

	runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(migrations.AddDocumentTextIndexStateQuery)
		return err
	})
	if runErr != nil {
		t.Fatalf("run migration: %v", runErr)
	}

	for _, row := range legacyRows {
		var state string
		var indexedBytes int64
		var failureReason sql.NullString
		var isStale bool
		queryErr := dbContext.QueryTx(func(tx *sql.Tx) error {
			return tx.QueryRow(
				`SELECT dt.state, dt.indexed_bytes, dt.failure_reason, dt.source_updated_at < hf.updated_at
				 FROM document_text dt JOIN home_file hf ON hf.id = dt.file_id WHERE hf.name = $1`, row.name,
			).Scan(&state, &indexedBytes, &failureReason, &isStale)
		})
		if queryErr != nil {
			t.Fatalf("read %s: %v", row.name, queryErr)
		}
		if state != row.wantState || indexedBytes != row.wantBytes || failureReason.String != row.wantReason || isStale != row.wantRequeued {
			t.Fatalf("%s: got state=%s bytes=%d reason=%q stale=%v, want state=%s bytes=%d reason=%q stale=%v",
				row.name, state, indexedBytes, failureReason.String, isStale, row.wantState, row.wantBytes, row.wantReason, row.wantRequeued)
		}
	}
}
