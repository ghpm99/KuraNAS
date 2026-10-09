package files

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

var searchSeedMoment = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func searchNames(t *testing.T, repo *Repository, query FileSearchQuery) (utils.PaginationResponse[FileModel], error) {
	t.Helper()
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 50
	}
	return repo.SearchActiveFiles(query)
}

func insertSearchRow(t *testing.T, repo *Repository, name string, path string, parentPath string, fileType FileType) {
	t.Helper()
	moment := searchSeedMoment
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

func TestPostgres_SearchGlobalIsCaseInsensitiveAndSkipsDeleted(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := searchNames(t, repo, FileSearchQuery{Query: "RELATORIO"})
	if err != nil {
		t.Fatalf("global search: %v", err)
	}
	expected := []string{"Relatorio.txt", "relatorio-final.txt", "relatorios", "relatorio-vizinho.txt"}
	if !slices.Equal(namesOf(page.Items), expected) {
		t.Fatalf("expected %v, got %v", expected, namesOf(page.Items))
	}
}

func TestPostgres_SearchUnderPathReturnsOnlyDescendants(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := searchNames(t, repo, FileSearchQuery{Query: "relatorio", Scope: SearchScopeDescendants, ScopePath: "/srv/docs/"})
	if err != nil {
		t.Fatalf("under path search: %v", err)
	}
	expected := []string{"Relatorio.txt", "relatorio-final.txt", "relatorios"}
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

	page, err := searchNames(t, repo, FileSearchQuery{Query: "nota", Scope: SearchScopeDescendants, ScopePath: `D:\Pasta\`})
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

	page, err := searchNames(t, repo, FileSearchQuery{Query: "relatorio", Scope: SearchScopeChildren, ScopePath: "/srv/docs"})
	if err != nil {
		t.Fatalf("children search: %v", err)
	}
	expected := []string{"Relatorio.txt", "relatorios"}
	if !slices.Equal(namesOf(page.Items), expected) {
		t.Fatalf("expected %v, got %v", expected, namesOf(page.Items))
	}
}

func TestPostgres_SearchTreatsPercentAndUnderscoreLiterally(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchTree(t, repo)

	page, err := searchNames(t, repo, FileSearchQuery{Query: "100%_"})
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

	firstPage, err := searchNames(t, repo, FileSearchQuery{Query: "relatorio", Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	secondPage, err := searchNames(t, repo, FileSearchQuery{Query: "relatorio", Page: 2, PageSize: 2})
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

type searchRowSpec struct {
	name         string
	path         string
	format       string
	fileType     FileType
	size         int64
	updatedAt    time.Time
	isStarred    bool
	physicalPath string
}

func seedSearchRows(t *testing.T, repo *Repository, rows []searchRowSpec) {
	t.Helper()
	truncateHomeFile(t, repo)
	for _, row := range rows {
		err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			created, createErr := repo.CreateFile(tx, FileModel{
				Name:       row.name,
				Path:       row.path,
				ParentPath: filepath.Dir(row.path),
				Format:     row.format,
				Size:       row.size,
				UpdatedAt:  row.updatedAt,
				CreatedAt:  row.updatedAt,
				Type:       row.fileType,
			})
			if createErr != nil {
				return createErr
			}
			_, updateErr := tx.Exec(
				"UPDATE home_file SET starred = $1, physical_path = NULLIF($2, '') WHERE id = $3",
				row.isStarred, row.physicalPath, created.ID,
			)
			return updateErr
		})
		if err != nil {
			t.Fatalf("seed %q: %v", row.path, err)
		}
	}
}

func searchFilteredNames(t *testing.T, repo *Repository, searchText string, filter FileSearchFilter) []string {
	t.Helper()
	page, err := searchNames(t, repo, FileSearchQuery{Query: searchText, Filter: filter})
	if err != nil {
		t.Fatalf("filtered search: %v", err)
	}
	return namesOf(page.Items)
}

func seedFilterRows(t *testing.T, repo *Repository) {
	t.Helper()
	january := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	february := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
	march := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	seedSearchRows(t, repo, []searchRowSpec{
		{name: "dados pasta", path: "/srv/dados pasta", fileType: Directory, updatedAt: january},
		{name: "dados.pdf", path: "/srv/dados.pdf", format: ".pdf", fileType: File, size: 500_000, updatedAt: january},
		{name: "dados.png", path: "/srv/dados.png", format: ".png", fileType: File, size: 5_000_000, updatedAt: february, isStarred: true},
		{name: "dados.mp3", path: "/srv/dados.mp3", format: ".mp3", fileType: File, size: 50_000_000, updatedAt: march},
		{name: "dados.mp4", path: "/srv/dados.mp4", format: ".mp4", fileType: File, size: 2_000_000_000, updatedAt: march, physicalPath: "/cold/dados.mp4"},
		{name: "dados.zip", path: "/srv/dados.zip", format: ".zip", fileType: File, size: 80_000, updatedAt: january},
		{name: "dados.xyz", path: "/srv/dados.xyz", format: ".xyz", fileType: File, size: 10, updatedAt: february},
	})
}

func sortedStrings(values []string) []string {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	return sorted
}

func TestPostgres_SearchFiltersByKind(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)

	tests := []struct {
		name     string
		kinds    []FileSearchKind
		expected []string
	}{
		{"folder", []FileSearchKind{SearchKindFolder}, []string{"dados pasta"}},
		{"document", []FileSearchKind{SearchKindDocument}, []string{"dados.pdf"}},
		{"image", []FileSearchKind{SearchKindImage}, []string{"dados.png"}},
		{"audio", []FileSearchKind{SearchKindAudio}, []string{"dados.mp3"}},
		{"video", []FileSearchKind{SearchKindVideo}, []string{"dados.mp4"}},
		{"archive", []FileSearchKind{SearchKindArchive}, []string{"dados.zip"}},
		{"other", []FileSearchKind{SearchKindOther}, []string{"dados.xyz"}},
		{"several", []FileSearchKind{SearchKindFolder, SearchKindImage, SearchKindOther}, []string{"dados pasta", "dados.png", "dados.xyz"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			names := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: tc.kinds})
			if !slices.Equal(sortedStrings(names), sortedStrings(tc.expected)) {
				t.Fatalf("expected %v, got %v", tc.expected, names)
			}
		})
	}
}

func TestPostgres_SearchFiltersByModifiedRange(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)
	february := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)

	fromNames := searchFilteredNames(t, repo, "dados", FileSearchFilter{ModifiedFrom: &february})
	if !slices.Equal(sortedStrings(fromNames), []string{"dados.mp3", "dados.mp4", "dados.png", "dados.xyz"}) {
		t.Fatalf("modified_from: got %v", fromNames)
	}
	toNames := searchFilteredNames(t, repo, "dados", FileSearchFilter{ModifiedTo: &february})
	if !slices.Equal(sortedStrings(toNames), []string{"dados pasta", "dados.pdf", "dados.png", "dados.xyz", "dados.zip"}) {
		t.Fatalf("modified_to must include the whole last day, got %v", toNames)
	}
}

func TestPostgres_SearchFiltersBySizeRange(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)
	minSize, maxSize := int64(1_000_000), int64(100_000_000)

	names := searchFilteredNames(t, repo, "dados", FileSearchFilter{MinSize: &minSize, MaxSize: &maxSize})
	if !slices.Equal(sortedStrings(names), []string{"dados.mp3", "dados.png"}) {
		t.Fatalf("size range: got %v", names)
	}
}

func TestPostgres_SearchFiltersByTierAndStarred(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)

	coldNames := searchFilteredNames(t, repo, "dados", FileSearchFilter{Tier: TierCold})
	if !slices.Equal(coldNames, []string{"dados.mp4"}) {
		t.Fatalf("cold tier: got %v", coldNames)
	}
	hotNames := searchFilteredNames(t, repo, "dados", FileSearchFilter{Tier: TierHot})
	if len(hotNames) != 6 || slices.Contains(hotNames, "dados.mp4") {
		t.Fatalf("hot tier: got %v", hotNames)
	}
	starredNames := searchFilteredNames(t, repo, "dados", FileSearchFilter{OnlyStarred: true})
	if !slices.Equal(starredNames, []string{"dados.png"}) {
		t.Fatalf("starred: got %v", starredNames)
	}
}

func TestPostgres_SearchCombinesFiltersWithAnd(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)
	minSize := int64(1_000_000)

	names := searchFilteredNames(t, repo, "dados", FileSearchFilter{
		Kinds:       []FileSearchKind{SearchKindImage, SearchKindVideo},
		MinSize:     &minSize,
		OnlyStarred: true,
	})
	if !slices.Equal(names, []string{"dados.png"}) {
		t.Fatalf("combined filters: got %v", names)
	}
}

func TestPostgres_SearchSortsByNameSizeAndModified(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedFilterRows(t, repo)
	fileOnly := []FileSearchKind{SearchKindDocument, SearchKindImage, SearchKindAudio}

	byNameAscending := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortName})
	if !slices.Equal(byNameAscending, []string{"dados.mp3", "dados.pdf", "dados.png"}) {
		t.Fatalf("name asc: got %v", byNameAscending)
	}
	byNameDescending := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortName, Order: SearchOrderDescending})
	if !slices.Equal(byNameDescending, []string{"dados.png", "dados.pdf", "dados.mp3"}) {
		t.Fatalf("name desc: got %v", byNameDescending)
	}
	bySizeDefault := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortSize})
	if !slices.Equal(bySizeDefault, []string{"dados.mp3", "dados.png", "dados.pdf"}) {
		t.Fatalf("size default (desc): got %v", bySizeDefault)
	}
	bySizeAscending := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortSize, Order: SearchOrderAscending})
	if !slices.Equal(bySizeAscending, []string{"dados.pdf", "dados.png", "dados.mp3"}) {
		t.Fatalf("size asc: got %v", bySizeAscending)
	}
	byModifiedDefault := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortModified})
	if !slices.Equal(byModifiedDefault, []string{"dados.mp3", "dados.png", "dados.pdf"}) {
		t.Fatalf("modified default (desc): got %v", byModifiedDefault)
	}
	byModifiedAscending := searchFilteredNames(t, repo, "dados", FileSearchFilter{Kinds: fileOnly, Sort: SearchSortModified, Order: SearchOrderAscending})
	if !slices.Equal(byModifiedAscending, []string{"dados.pdf", "dados.png", "dados.mp3"}) {
		t.Fatalf("modified asc: got %v", byModifiedAscending)
	}
}

func TestPostgres_SearchRelevanceRanksExactThenPrefixThenSubstringThenStarredThenRecent(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	seedSearchRows(t, repo, []searchRowSpec{
		{name: "meu relatorio final", path: "/srv/a/meu relatorio final", format: ".txt", fileType: File, updatedAt: newer, isStarred: true},
		{name: "relatorio antigo", path: "/srv/b/relatorio antigo", format: ".txt", fileType: File, updatedAt: older},
		{name: "relatorio novo", path: "/srv/c/relatorio novo", format: ".txt", fileType: File, updatedAt: newer},
		{name: "relatorio favorito", path: "/srv/d/relatorio favorito", format: ".txt", fileType: File, updatedAt: older, isStarred: true},
		{name: "Relatorio", path: "/srv/e/Relatorio", format: ".txt", fileType: File, updatedAt: older},
		{name: "outro relatorio", path: "/srv/f/outro relatorio", format: ".txt", fileType: File, updatedAt: older},
	})

	names := searchFilteredNames(t, repo, "relatorio", FileSearchFilter{})
	expected := []string{
		"Relatorio",
		"relatorio favorito",
		"relatorio novo",
		"relatorio antigo",
		"meu relatorio final",
		"outro relatorio",
	}
	if !slices.Equal(names, expected) {
		t.Fatalf("expected %v, got %v", expected, names)
	}
}

func TestPostgres_SearchMatchesEveryTermInAnyOrder(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	seedSearchRows(t, repo, []searchRowSpec{
		{name: "relatorio anual 2024.pdf", path: "/srv/relatorio anual 2024.pdf", format: ".pdf", fileType: File, updatedAt: searchSeedMoment},
		{name: "2024 anual relatorio.pdf", path: "/srv/2024 anual relatorio.pdf", format: ".pdf", fileType: File, updatedAt: searchSeedMoment},
		{name: "relatorio mensal 2024.pdf", path: "/srv/relatorio mensal 2024.pdf", format: ".pdf", fileType: File, updatedAt: searchSeedMoment},
	})

	names := searchFilteredNames(t, repo, "  Anual   relatorio ", FileSearchFilter{})
	if !slices.Equal(sortedStrings(names), []string{"2024 anual relatorio.pdf", "relatorio anual 2024.pdf"}) {
		t.Fatalf("multi-word AND match: got %v", names)
	}
}
