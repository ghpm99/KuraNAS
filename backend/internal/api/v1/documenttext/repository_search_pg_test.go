package documenttext

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

func nameOnlyDocument(t *testing.T, repository *Repository, fileID int, sourceUpdatedAt time.Time) {
	t.Helper()
	err := repository.UpsertDocumentText(DocumentTextModel{
		FileID: fileID, SourceUpdatedAt: sourceUpdatedAt, State: StateNameOnly, FailureReason: "scanned",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func nineStoredPages(termOnPage int, term string) string {
	pages := make([]string, 0, 9)
	for pageNumber := 1; pageNumber <= 9; pageNumber++ {
		pageText := "texto corrido da pagina " + strings.Repeat("palavra ", 30)
		if pageNumber == termOnPage {
			pageText += " aqui aparece o " + term + " procurado "
		}
		pages = append(pages, pageText)
	}
	return strings.Join(pages, "\f")
}

func searchPage(t *testing.T, service *Service, params DocumentSearchParams) []DocumentSearchResultDto {
	t.Helper()
	if params.Page == 0 {
		params.Page, params.PageSize = 1, 50
	}
	response, err := service.SearchDocuments(params)
	if err != nil {
		t.Fatal(err)
	}
	return response.Items
}

func TestSearchDocumentsReportsPageAndSnippetOfContentMatch_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fileID := seedDocumentFile(t, dbContext, "manual.pdf", ".pdf", now)
	storeText(t, repository, fileID, nineStoredPages(7, "Reembolso"), now, "")

	results := searchPage(t, service, DocumentSearchParams{Query: "reembolso"})

	if len(results) != 1 || results[0].Page == nil || *results[0].Page != 7 {
		t.Fatalf("expected page 7, got %+v", results)
	}
	if !strings.Contains(results[0].Snippet, "Reembolso") || strings.Contains(results[0].Snippet, "\f") || len(results[0].Snippet) > 220 {
		t.Fatalf("unexpected snippet %q", results[0].Snippet)
	}
}

func TestSearchDocumentsPageFollowsTheTermThatAppearsLast_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fileID := seedDocumentFile(t, dbContext, "manual.pdf", ".pdf", now)
	pages := make([]string, 0, 9)
	for pageNumber := 1; pageNumber <= 9; pageNumber++ {
		pages = append(pages, fmt.Sprintf("secao comum da pagina %d", pageNumber))
	}
	pages[5] = "secao especial 1799 da pagina 6"
	storeText(t, repository, fileID, strings.Join(pages, "\f"), now, "")

	results := searchPage(t, service, DocumentSearchParams{Query: "secao 1799"})

	if len(results) != 1 || results[0].Page == nil || *results[0].Page != 6 {
		t.Fatalf("expected page 6, got %+v", results)
	}
	if !strings.Contains(results[0].Snippet, "1799") {
		t.Fatalf("expected the snippet around 1799, got %q", results[0].Snippet)
	}
}

func TestSearchDocumentsPageIsNullWithoutPageBreaks_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fileID := seedDocumentFile(t, dbContext, "nota.txt", ".txt", now)
	storeText(t, repository, fileID, "uma nota com o termo reembolso no meio", now, "")

	results := searchPage(t, service, DocumentSearchParams{Query: "reembolso"})

	if len(results) != 1 || results[0].Page != nil || results[0].Snippet == "" {
		t.Fatalf("expected snippet and null page, got %+v", results)
	}
}

func TestSearchDocumentsPageOfFirstPageMatchIsOne_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fileID := seedDocumentFile(t, dbContext, "manual.pdf", ".pdf", now)
	storeText(t, repository, fileID, nineStoredPages(1, "Reembolso"), now, "")

	results := searchPage(t, service, DocumentSearchParams{Query: "reembolso"})

	if len(results) != 1 || results[0].Page == nil || *results[0].Page != 1 {
		t.Fatalf("expected page 1, got %+v", results)
	}
}

func TestSearchDocumentsFindsNameOnlyDocumentByNameWithoutSnippet_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	scanned := seedDocumentFile(t, dbContext, "Contrato Digitalizado.pdf", ".pdf", now)
	nameOnlyDocument(t, repository, scanned, now)
	neverIndexed := seedDocumentFile(t, dbContext, "contrato fila.pdf", ".pdf", now.Add(-time.Hour))

	results := searchPage(t, service, DocumentSearchParams{Query: "contrato"})

	if got := resultNames(results); !slices.Equal(got, []string{"Contrato Digitalizado.pdf", "contrato fila.pdf"}) {
		t.Fatalf("unexpected names %v", got)
	}
	for _, result := range results {
		if result.Snippet != "" || result.Page != nil {
			t.Fatalf("name-only match must have empty snippet and null page: %+v", result)
		}
	}
	if results[1].FileID != neverIndexed {
		t.Fatalf("unexpected second result %+v", results[1])
	}
}

func TestSearchDocumentsFailedDocumentMatchesOnlyByName_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	failed := seedDocumentFile(t, dbContext, "relatorio.docx", ".docx", now)
	storeText(t, repository, failed, "segredo", now, "binary")

	if results := searchPage(t, service, DocumentSearchParams{Query: "segredo"}); len(results) != 0 {
		t.Fatalf("failed text must not be searched, got %v", resultNames(results))
	}
	if results := searchPage(t, service, DocumentSearchParams{Query: "relatorio"}); len(results) != 1 || results[0].Snippet != "" {
		t.Fatalf("expected name match without snippet, got %+v", results)
	}
}

func TestSearchDocumentsRanksNameThenContentThenRecencyThenID_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	contentOld := seedDocumentFile(t, dbContext, "aaa.txt", ".txt", now.Add(-3*time.Hour))
	contentNew := seedDocumentFile(t, dbContext, "bbb.txt", ".txt", now)
	nameOldFirst := seedDocumentFile(t, dbContext, "orcamento um.txt", ".txt", now.Add(-24*time.Hour))
	nameOldSecond := seedDocumentFile(t, dbContext, "orcamento dois.txt", ".txt", now.Add(-24*time.Hour))
	for _, fileID := range []int{contentOld, contentNew} {
		storeText(t, repository, fileID, "revisao do orcamento", now, "")
	}
	nameOnlyDocument(t, repository, nameOldFirst, now)
	nameOnlyDocument(t, repository, nameOldSecond, now)

	results := searchPage(t, service, DocumentSearchParams{Query: "orcamento"})

	want := []string{"orcamento um.txt", "orcamento dois.txt", "bbb.txt", "aaa.txt"}
	if got := resultNames(results); !slices.Equal(got, want) {
		t.Fatalf("ranking = %v, want %v", got, want)
	}
}

func TestSearchDocumentsAppliesLibraryFilters_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	folderID := seedLibraryEntry(t, dbContext, libraryEntry{name: "projetos", directory: "/docs", updatedAt: now, isFolder: true})
	inside := seedLibraryEntry(t, dbContext, libraryEntry{name: "plano.pdf", directory: "/docs/projetos", size: 10, updatedAt: now, starred: true})
	outside := seedLibraryEntry(t, dbContext, libraryEntry{name: "plano.txt", directory: "/docs/outros", size: 10, updatedAt: now.Add(-72 * time.Hour)})
	storeText(t, repository, inside, "plano de projeto", now, "")
	storeText(t, repository, outside, "plano de projeto", now.Add(-72*time.Hour), "")

	filters := map[string]DocumentSearchParams{
		"format":        {Filter: DocumentFilter{Format: GroupPDF}},
		"folder":        {FolderID: folderID},
		"favorites":     {Filter: DocumentFilter{FavoritesOnly: true}},
		"modified from": {Filter: DocumentFilter{ModifiedFrom: &now}},
	}
	for name, params := range filters {
		params.Query, params.Page, params.PageSize = "plano", 1, 50
		results := searchPage(t, service, params)
		if len(results) != 1 || results[0].FileID != inside {
			t.Fatalf("%s: expected only %d, got %v", name, inside, resultNames(results))
		}
		if !results[0].Starred {
			t.Fatalf("%s: expected the starred document to come back starred", name)
		}
	}

	oldDate := now.Add(-96 * time.Hour)
	untilYesterday := now.Add(-48 * time.Hour)
	results := searchPage(t, service, DocumentSearchParams{Query: "plano", Page: 1, PageSize: 50, Filter: DocumentFilter{ModifiedFrom: &oldDate, ModifiedTo: &untilYesterday}})
	if len(results) != 1 || results[0].FileID != outside {
		t.Fatalf("modified range: expected only %d, got %v", outside, resultNames(results))
	}
}

func TestSearchDocumentsFiltersByIndexState_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	indexed := seedDocumentFile(t, dbContext, "alvo indexado.pdf", ".pdf", now)
	scanned := seedDocumentFile(t, dbContext, "alvo digitalizado.pdf", ".pdf", now)
	storeText(t, repository, indexed, "conteudo", now, "")
	nameOnlyDocument(t, repository, scanned, now)

	results := searchPage(t, service, DocumentSearchParams{Query: "alvo", Filter: DocumentFilter{IndexState: StateNameOnly}})

	if len(results) != 1 || results[0].FileID != scanned {
		t.Fatalf("expected only the name_only document, got %v", resultNames(results))
	}
}

func TestSearchDocumentsUnknownFolderIsNotFound_Postgres(t *testing.T) {
	_, service, _ := newDocumentTextPostgres(t)

	_, err := service.SearchDocuments(DocumentSearchParams{Query: "plano", FolderID: 999, Page: 1, PageSize: 10})

	if err != ErrFolderNotFound {
		t.Fatalf("expected ErrFolderNotFound, got %v", err)
	}
}

func TestSearchDocumentsExcludesTrashAndNonDocumentFormats_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	kept := seedDocumentFile(t, dbContext, "alvo.txt", ".txt", now)
	trashed := seedDocumentFile(t, dbContext, "alvo lixeira.txt", ".txt", now)
	sourceCode := seedDocumentFile(t, dbContext, "alvo.go", ".go", now)
	logFile := seedDocumentFile(t, dbContext, "alvo.log", ".log", now)
	for _, fileID := range []int{kept, trashed, sourceCode, logFile} {
		storeText(t, repository, fileID, "alvo no conteudo", now, "")
	}
	if err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET deleted_at = now() WHERE id = $1`, trashed)
		return err
	}); err != nil {
		t.Fatal(err)
	}

	if got := resultNames(searchPage(t, service, DocumentSearchParams{Query: "alvo"})); !slices.Equal(got, []string{"alvo.txt"}) {
		t.Fatalf("unexpected names %v", got)
	}
	topResults, err := service.SearchTopDocuments("alvo", 10)
	if err != nil || !slices.Equal(resultNames(topResults), []string{"alvo.txt"}) {
		t.Fatalf("global search names = %v err=%v", resultNames(topResults), err)
	}
}
