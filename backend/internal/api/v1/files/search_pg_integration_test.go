package files

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func insertSearchRow(t *testing.T, repo *Repository, name string, path string, parentPath string, fileType FileType) {
	t.Helper()
	moment := time.Now().UTC().Truncate(time.Second)
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, createErr := repo.CreateFile(tx, FileModel{
			Name:       name,
			Path:       path,
			ParentPath: parentPath,
			Format:     ".txt",
			UpdatedAt:  moment,
			CreatedAt:  moment,
			Type:       fileType,
		})
		return createErr
	})
	if err != nil {
		t.Fatalf("insert %q: %v", path, err)
	}
}

func namesOf(models []FileModel) []string {
	names := make([]string, 0, len(models))
	for _, model := range models {
		names = append(names, model.Name)
	}
	return names
}

func seedSearchTree(t *testing.T, repo *Repository) {
	t.Helper()
	truncateHomeFile(t, repo)
	insertSearchRow(t, repo, "docs", "/srv/docs", "/srv", Directory)
	insertSearchRow(t, repo, "docs2", "/srv/docs2", "/srv", Directory)
	insertSearchRow(t, repo, "Relatorio.txt", "/srv/docs/Relatorio.txt", "/srv/docs", File)
	insertSearchRow(t, repo, "relatorio-final.txt", "/srv/docs/2024/relatorio-final.txt", "/srv/docs/2024", File)
	insertSearchRow(t, repo, "relatorios", "/srv/docs/relatorios", "/srv/docs", Directory)
	insertSearchRow(t, repo, "relatorio-vizinho.txt", "/srv/docs2/relatorio-vizinho.txt", "/srv/docs2", File)
	insertSearchRow(t, repo, "relatorio-apagado.txt", "/srv/docs/relatorio-apagado.txt", "/srv/docs", File)
	insertSearchRow(t, repo, "100%_real.txt", "/srv/docs/100%_real.txt", "/srv/docs", File)
	insertSearchRow(t, repo, "100xyreal.txt", "/srv/docs/100xyreal.txt", "/srv/docs", File)

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET deleted_at = now() WHERE name = 'relatorio-apagado.txt'")
		return execErr
	})
	if err != nil {
		t.Fatalf("soft delete: %v", err)
	}
}

func TestPostgres_SearchGlobalIsCaseInsensitiveSkipsDeletedAndListsDirectoriesFirst(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := repo.SearchActiveFilesByName(utils.BuildContainsLikePattern("RELATORIO"), 1, 50)
	if err != nil {
		t.Fatalf("global search: %v", err)
	}
	expected := []string{"relatorios", "Relatorio.txt", "relatorio-final.txt", "relatorio-vizinho.txt"}
	if !slices.Equal(namesOf(page.Items), expected) {
		t.Fatalf("expected %v, got %v", expected, namesOf(page.Items))
	}
}

func TestPostgres_SearchUnderPathReturnsOnlyDescendants(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := repo.SearchActiveFilesByNameUnderPath("/srv/docs/", utils.BuildContainsLikePattern("relatorio"), 1, 50)
	if err != nil {
		t.Fatalf("under path search: %v", err)
	}
	expected := []string{"relatorios", "Relatorio.txt", "relatorio-final.txt"}
	if !slices.Equal(namesOf(page.Items), expected) {
		t.Fatalf("expected %v, got %v", expected, namesOf(page.Items))
	}
}

func TestPostgres_SearchUnderPathMatchesWindowsPaths(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)
	insertSearchRow(t, repo, "nota.txt", `D:\Pasta\sub\nota.txt`, `D:\Pasta\sub`, File)
	insertSearchRow(t, repo, "nota.txt", `D:\Pasta2\nota.txt`, `D:\Pasta2`, File)

	page, err := repo.SearchActiveFilesByNameUnderPath(`D:\Pasta\`, utils.BuildContainsLikePattern("nota"), 1, 50)
	if err != nil {
		t.Fatalf("windows search: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Path != `D:\Pasta\sub\nota.txt` {
		t.Fatalf("expected only the Pasta descendant, got %+v", page.Items)
	}
}

func TestPostgres_SearchChildrenReturnsOnlyDirectChildren(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := repo.SearchActiveChildrenByName("/srv/docs", utils.BuildContainsLikePattern("relatorio"), 1, 50)
	if err != nil {
		t.Fatalf("children search: %v", err)
	}
	expected := []string{"relatorios", "Relatorio.txt"}
	if !slices.Equal(namesOf(page.Items), expected) {
		t.Fatalf("expected %v, got %v", expected, namesOf(page.Items))
	}
}

func TestPostgres_SearchTreatsPercentAndUnderscoreLiterally(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := repo.SearchActiveFilesByName(utils.BuildContainsLikePattern("100%_"), 1, 50)
	if err != nil {
		t.Fatalf("escaped search: %v", err)
	}
	if !slices.Equal(namesOf(page.Items), []string{"100%_real.txt"}) {
		t.Fatalf("wildcards must be literal, got %v", namesOf(page.Items))
	}
}

func TestPostgres_SearchPaginatesCompleteResults(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	firstPage, err := repo.SearchActiveFilesByName(utils.BuildContainsLikePattern("relatorio"), 1, 2)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	secondPage, err := repo.SearchActiveFilesByName(utils.BuildContainsLikePattern("relatorio"), 2, 2)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(firstPage.Items) != 2 || !firstPage.Pagination.HasNext {
		t.Fatalf("first page must be full with next, got %+v", firstPage.Pagination)
	}
	if len(secondPage.Items) != 2 || secondPage.Pagination.HasNext {
		t.Fatalf("second page must hold the rest, got %d items %+v", len(secondPage.Items), secondPage.Pagination)
	}
}

func TestPostgres_NameTrigramIndexExistsWhenExtensionIsAvailable(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)

	var hasExtension bool
	var hasIndex bool
	err := repo.GetDbContext().QueryTx(func(tx *sql.Tx) error {
		if scanErr := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')").Scan(&hasExtension); scanErr != nil {
			return scanErr
		}
		return tx.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'home_file_active_name_trigram')").Scan(&hasIndex)
	})
	if err != nil {
		t.Fatalf("inspect index: %v", err)
	}
	if hasExtension && !hasIndex {
		t.Fatalf("pg_trgm is installed but the name trigram index is missing")
	}
}

func TestPostgres_FolderStatsCountsOnlyActiveDescendantsOfTheFolder(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)
	sizeErr := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET size = 10 WHERE type = 2")
		return execErr
	})
	if sizeErr != nil {
		t.Fatalf("set sizes: %v", sizeErr)
	}

	stats, err := repo.GetFolderStats("/srv/docs/")
	if err != nil {
		t.Fatalf("folder stats: %v", err)
	}
	if stats.FileCount != 4 || stats.FolderCount != 1 || stats.TotalSizeBytes != 40 {
		t.Fatalf("expected 4 files, 1 folder and 40 bytes under /srv/docs (soft-deleted and sibling docs2 excluded), got %+v", stats)
	}

	emptyStats, err := repo.GetFolderStats("/srv/nothing/")
	if err != nil {
		t.Fatalf("empty folder stats: %v", err)
	}
	if emptyStats != (FolderStatsDto{}) {
		t.Fatalf("expected zero stats for an unknown folder, got %+v", emptyStats)
	}
}
