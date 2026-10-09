package documenttext

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

func newDocumentTextPostgres(t *testing.T) (*Repository, *Service, *database.DbContext) {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_documenttext_it")
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`TRUNCATE document_text, home_file RESTART IDENTITY CASCADE`)
		return err
	})
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	repository := NewRepository(dbContext)
	return repository, &Service{Repository: repository}, dbContext
}

func seedDocumentFile(t *testing.T, dbContext *database.DbContext, name string, format string, updatedAt time.Time) int {
	t.Helper()
	var fileID int
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			 VALUES ($1::text, '/docs/' || $1::text, '/docs', $2::text, 100, $3::timestamptz, $3::timestamptz, 2, '') RETURNING id`,
			name, format, updatedAt,
		).Scan(&fileID)
	})
	if err != nil {
		t.Fatalf("seed file: %v", err)
	}
	return fileID
}

func storeText(t *testing.T, repository *Repository, fileID int, text string, sourceUpdatedAt time.Time, errorCode string) {
	t.Helper()
	err := repository.UpsertDocumentText(DocumentTextModel{
		FileID: fileID, ExtractedText: text, TextLength: len([]rune(text)), SourceUpdatedAt: sourceUpdatedAt, ErrorCode: errorCode,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
}

func resultNames(results []DocumentSearchResultDto) []string {
	names := make([]string, 0, len(results))
	for _, result := range results {
		names = append(names, result.Name)
	}
	return names
}

func TestSearchDocumentsMatchesContentAccentInsensitively_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	contract := seedDocumentFile(t, dbContext, "contrato.txt", ".txt", now)
	other := seedDocumentFile(t, dbContext, "outro.md", ".md", now)
	storeText(t, repository, contract, "Cláusula de rescisão do contrato de locação residencial", now, "")
	storeText(t, repository, other, "Lista de compras: pão e leite", now, "")

	for _, query := range []string{"rescisao", "RESCISÃO", "clausula locacao"} {
		results, err := service.SearchTopDocuments(query, 10)
		if err != nil {
			t.Fatalf("search %q: %v", query, err)
		}
		if !slices.Equal(resultNames(results), []string{"contrato.txt"}) {
			t.Fatalf("query %q names = %v", query, resultNames(results))
		}
	}

	results, err := service.SearchTopDocuments("rescisao", 10)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].FileID != contract || results[0].Snippet != "Cláusula de rescisão do contrato de locação residencial" || results[0].Format != ".txt" || results[0].Size != 100 {
		t.Fatalf("unexpected result %+v", results[0])
	}
}

func TestSearchDocumentsRequiresEveryTerm_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	fileID := seedDocumentFile(t, dbContext, "a.txt", ".txt", now)
	storeText(t, repository, fileID, "alpha beta gamma", now, "")

	results, err := service.SearchTopDocuments("alpha delta", 10)
	if err != nil || len(results) != 0 {
		t.Fatalf("expected no results, got %v err=%v", resultNames(results), err)
	}
}

func TestSearchDocumentsRanksNameMatchFirstThenRecency_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	oldNameMatch := seedDocumentFile(t, dbContext, "orcamento anual.txt", ".txt", now.Add(-48*time.Hour))
	recentContent := seedDocumentFile(t, dbContext, "notas.txt", ".txt", now)
	olderContent := seedDocumentFile(t, dbContext, "diario.txt", ".txt", now.Add(-time.Hour))
	for _, fileID := range []int{oldNameMatch, recentContent, olderContent} {
		storeText(t, repository, fileID, "revisao do orcamento do ano", now, "")
	}

	results, err := service.SearchTopDocuments("orçamento", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := resultNames(results); !slices.Equal(got, []string{"orcamento anual.txt", "notas.txt", "diario.txt"}) {
		t.Fatalf("unexpected ranking %v", got)
	}
}

func TestSearchDocumentsPagination_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	for index, name := range []string{"a.txt", "b.txt", "c.txt"} {
		fileID := seedDocumentFile(t, dbContext, name, ".txt", now.Add(-time.Duration(index)*time.Hour))
		storeText(t, repository, fileID, "termo comum", now, "")
	}

	firstPage, err := service.SearchDocuments("termo", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	secondPage, err := service.SearchDocuments("termo", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(resultNames(firstPage.Items), []string{"a.txt", "b.txt"}) || !firstPage.Pagination.HasNext {
		t.Fatalf("unexpected first page %+v", firstPage)
	}
	if !slices.Equal(resultNames(secondPage.Items), []string{"c.txt"}) || secondPage.Pagination.HasNext || !secondPage.Pagination.HasPrev {
		t.Fatalf("unexpected second page %+v", secondPage)
	}
}

func TestSearchDocumentsIgnoresErrorRowsAndDeletedFiles_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	failed := seedDocumentFile(t, dbContext, "failed.txt", ".txt", now)
	deleted := seedDocumentFile(t, dbContext, "deleted.txt", ".txt", now)
	storeText(t, repository, failed, "segredo", now, "binary")
	storeText(t, repository, deleted, "segredo", now, "")
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET deleted_at = now() WHERE id = $1`, deleted)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := service.SearchTopDocuments("segredo", 10)
	if err != nil || len(results) != 0 {
		t.Fatalf("expected no results, got %v err=%v", resultNames(results), err)
	}
}

func TestSearchDocumentsEscapesLikeWildcards_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	fileID := seedDocumentFile(t, dbContext, "a.txt", ".txt", now)
	storeText(t, repository, fileID, "relatorio simples", now, "")

	results, err := service.SearchTopDocuments("100%", 10)
	if err != nil || len(results) != 0 {
		t.Fatalf("wildcard must be literal, got %v err=%v", resultNames(results), err)
	}
}

func TestListPendingIndexingTracksMissingAndStaleRows_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fresh := seedDocumentFile(t, dbContext, "fresh.txt", ".txt", now)
	missing := seedDocumentFile(t, dbContext, "missing.pdf", ".pdf", now)
	stale := seedDocumentFile(t, dbContext, "stale.md", ".md", now)
	seedDocumentFile(t, dbContext, "photo.jpg", ".jpg", now)
	deleted := seedDocumentFile(t, dbContext, "deleted.txt", ".txt", now)
	storeText(t, repository, fresh, "x", now, "")
	storeText(t, repository, stale, "old", now.Add(-time.Hour), "")
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET deleted_at = now() WHERE id = $1`, deleted)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	pending, err := repository.ListPendingIndexing(0, 10)
	if err != nil {
		t.Fatal(err)
	}
	pendingIDs := []int{}
	for _, pendingDocument := range pending {
		pendingIDs = append(pendingIDs, pendingDocument.FileID)
	}
	if !slices.Equal(pendingIDs, []int{missing, stale}) {
		t.Fatalf("unexpected pending ids %v", pendingIDs)
	}
	if pending[0].Path != "/docs/missing.pdf" || pending[0].Format != ".pdf" || pending[0].Size != 100 {
		t.Fatalf("unexpected pending row %+v", pending[0])
	}

	afterFirst, err := repository.ListPendingIndexing(missing, 10)
	if err != nil || len(afterFirst) != 1 || afterFirst[0].FileID != stale {
		t.Fatalf("keyset cursor failed: %+v err=%v", afterFirst, err)
	}
}

func TestReindexAfterFileUpdateReplacesStoredText_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	indexedAt := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	fileID := seedDocumentFile(t, dbContext, "living.txt", ".txt", indexedAt)
	storeText(t, repository, fileID, "versao antiga", indexedAt, "")

	pending, err := repository.ListPendingIndexing(0, 10)
	if err != nil || len(pending) != 0 {
		t.Fatalf("indexed file must not be pending: %+v err=%v", pending, err)
	}

	updatedAt := indexedAt.Add(30 * time.Minute)
	err = dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET updated_at = $2 WHERE id = $1`, fileID, updatedAt)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	pending, err = repository.ListPendingIndexing(0, 10)
	if err != nil || len(pending) != 1 || pending[0].FileID != fileID {
		t.Fatalf("updated file must be pending again: %+v err=%v", pending, err)
	}

	storeText(t, repository, fileID, "versao nova", pending[0].UpdatedAt, "")

	oldResults, _ := service.SearchTopDocuments("antiga", 10)
	newResults, _ := service.SearchTopDocuments("nova", 10)
	if len(oldResults) != 0 || len(newResults) != 1 {
		t.Fatalf("reindex did not replace text: old=%v new=%v", resultNames(oldResults), resultNames(newResults))
	}
	stillPending, err := repository.ListPendingIndexing(0, 10)
	if err != nil || len(stillPending) != 0 {
		t.Fatalf("reindexed file must not be pending: %+v err=%v", stillPending, err)
	}
}

func TestDocumentTextRowIsDeletedWithHomeFile_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	fileID := seedDocumentFile(t, dbContext, "gone.txt", ".txt", now)
	storeText(t, repository, fileID, "texto", now, "")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`DELETE FROM home_file WHERE id = $1`, fileID)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	var remainingRows int
	err = dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT COUNT(*) FROM document_text WHERE file_id = $1`, fileID).Scan(&remainingRows)
	})
	if err != nil || remainingRows != 0 {
		t.Fatalf("expected cascade delete, rows=%d err=%v", remainingRows, err)
	}
}

func TestUpsertStoresErrorCodeAsNullWhenEmpty_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	okFile := seedDocumentFile(t, dbContext, "ok.txt", ".txt", now)
	badFile := seedDocumentFile(t, dbContext, "bad.txt", ".txt", now)
	storeText(t, repository, okFile, "texto", now, "")
	storeText(t, repository, badFile, "", now, "too_large")

	var okError, badError sql.NullString
	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT error FROM document_text WHERE file_id = $1`, okFile).Scan(&okError); err != nil {
			return err
		}
		return tx.QueryRow(`SELECT error FROM document_text WHERE file_id = $1`, badFile).Scan(&badError)
	})
	if err != nil {
		t.Fatal(err)
	}
	if okError.Valid || !badError.Valid || badError.String != "too_large" {
		t.Fatalf("unexpected error columns ok=%+v bad=%+v", okError, badError)
	}
}

func TestDocumentTextIndexExists_Postgres(t *testing.T) {
	_, _, dbContext := newDocumentTextPostgres(t)

	var hasTrigram bool
	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')`).Scan(&hasTrigram)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasTrigram {
		t.Skip("pg_trgm unavailable; the index is created defensively and skipped")
	}

	var hasIndex bool
	err = dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_document_text_extracted_text_fold_trigram')`).Scan(&hasIndex)
	})
	if err != nil || !hasIndex {
		t.Fatalf("trigram index missing: has=%v err=%v", hasIndex, err)
	}
}
