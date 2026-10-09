package search

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"nas-go/api/pkg/database/migrations"
)

func TestSearchFilesIgnoresAccentsInBothDirections_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "Relatório Anual.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "relatorio simples.txt", format: ".txt", fileType: 2, updatedAt: now.Add(-time.Hour)})
	seedHomeFile(t, dbContext, searchSeedFile{name: "outro.txt", format: ".txt", fileType: 2, updatedAt: now})

	for _, query := range []string{"relatorio", "RELATÓRIO", "relatório anual"} {
		results, err := repository.SearchFiles(query, 10)
		if err != nil {
			t.Fatalf("SearchFiles(%q): %v", query, err)
		}
		if query == "relatório anual" {
			if got := fileNames(results); !slices.Equal(got, []string{"Relatório Anual.txt"}) {
				t.Fatalf("query %q names = %v", query, got)
			}
			continue
		}
		if len(results) != 2 {
			t.Fatalf("query %q expected both report files, got %v", query, fileNames(results))
		}
	}
}

func TestSearchTracksIgnoresAccentsInTitleAndArtist_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	byTitle := seedHomeFile(t, dbContext, searchSeedFile{name: "01.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	byArtist := seedHomeFile(t, dbContext, searchSeedFile{name: "02.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedAudioTrackMetadata(t, dbContext, byTitle, "Música Popular", "Someone", "Other")
	seedAudioTrackMetadata(t, dbContext, byArtist, "Faixa", "Beyoncé", "Other")

	titleResults, err := repository.SearchTracks("musica", 10)
	if err != nil {
		t.Fatalf("SearchTracks: %v", err)
	}
	if got := trackTitles(titleResults); !slices.Equal(got, []string{"Música Popular"}) {
		t.Fatalf("titles = %v", got)
	}

	artistResults, err := repository.SearchTracks("beyonce", 10)
	if err != nil {
		t.Fatalf("SearchTracks: %v", err)
	}
	if got := trackTitles(artistResults); !slices.Equal(got, []string{"Faixa"}) {
		t.Fatalf("titles = %v", got)
	}
}

func TestSearchImagesIgnoresAccentsInCaption_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	imageID := seedHomeFile(t, dbContext, searchSeedFile{name: "IMG_1.jpg", format: ".jpg", fileType: 2, updatedAt: now})
	seedImageMetadata(t, dbContext, imageID, "praia ensolarada com coração")

	results, err := repository.SearchImages("CORACAO", 10)
	if err != nil {
		t.Fatalf("SearchImages: %v", err)
	}
	if len(results) != 1 || results[0].ID != imageID {
		t.Fatalf("results = %+v", results)
	}
}

func TestFuzzyFallbackFindsTypoedFileAndFolderNames_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "relatorio.pdf", format: ".pdf", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "fotografia.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "Relatórios", format: "", fileType: 1, updatedAt: now})

	if !repository.IsFuzzySearchAvailable() {
		t.Skip("pg_trgm unavailable")
	}

	files, err := repository.SearchFilesFuzzy("relatoro", 10)
	if err != nil {
		t.Fatalf("SearchFilesFuzzy: %v", err)
	}
	if got := fileNames(files); !slices.Equal(got, []string{"relatorio.pdf"}) {
		t.Fatalf("files = %v", got)
	}

	folders, err := repository.SearchFoldersFuzzy("relatorio", 10)
	if err != nil {
		t.Fatalf("SearchFoldersFuzzy: %v", err)
	}
	if len(folders) != 1 || folders[0].Name != "Relatórios" {
		t.Fatalf("folders = %+v", folders)
	}

	service := NewService(repository, nil)
	response, err := service.SearchGlobal("relatoro", 6)
	if err != nil {
		t.Fatalf("SearchGlobal: %v", err)
	}
	if !response.Fuzzy || len(response.Files) != 1 {
		t.Fatalf("expected fuzzy response with one file, got %+v", response)
	}

	exact, err := service.SearchGlobal("relatorio", 6)
	if err != nil {
		t.Fatalf("SearchGlobal: %v", err)
	}
	if exact.Fuzzy || len(exact.Files) != 1 {
		t.Fatalf("exact match must not be fuzzy, got %+v", exact)
	}
}

func TestFoldFunctionsExistAndFoldAccents_Postgres(t *testing.T) {
	_, dbContext := newSearchPostgresRepository(t)
	var foldedName string
	var foldedTerms string
	err := dbContext.GetDatabase().QueryRow(`SELECT kuranas_fold('Relatório'), array_to_string(kuranas_fold_terms(ARRAY['%Ação%', 'Ü']), ',')`).Scan(&foldedName, &foldedTerms)
	if err != nil {
		t.Fatalf("fold functions: %v", err)
	}
	if foldedName != "relatorio" || foldedTerms != "%acao%,u" {
		t.Fatalf("folded = %q / %q", foldedName, foldedTerms)
	}
}

func TestFoldFunctionFallsBackToLowercaseWhenExtensionsCannotBeCreated_Postgres(t *testing.T) {
	_, adminContext := newSearchPostgresRepository(t)
	admin := adminContext.GetDatabase()
	const restrictedRole = "kuranas_fold_restricted_it"
	const restrictedDatabase = "kuranas_fold_restricted_it"

	cleanup := func() {
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + restrictedDatabase)
		_, _ = admin.Exec(`DROP ROLE IF EXISTS ` + restrictedRole)
	}
	cleanup()
	t.Cleanup(cleanup)

	setupStatements := []string{
		`CREATE ROLE ` + restrictedRole + ` LOGIN PASSWORD 'restricted' NOSUPERUSER`,
		`CREATE DATABASE ` + restrictedDatabase + ` OWNER ` + restrictedRole,
		`REVOKE CREATE ON DATABASE ` + restrictedDatabase + ` FROM ` + restrictedRole,
	}
	for _, statement := range setupStatements {
		if _, err := admin.Exec(statement); err != nil {
			t.Skipf("cannot prepare restricted database: %v", err)
		}
	}

	restricted, err := sql.Open("postgres", restrictedConnectionString(restrictedRole, restrictedDatabase))
	if err != nil {
		t.Skipf("cannot open restricted database: %v", err)
	}
	defer restricted.Close()
	if pingErr := restricted.Ping(); pingErr != nil {
		t.Skipf("cannot reach restricted database: %v", pingErr)
	}

	if _, err := restricted.Exec(migrations.CreateSearchFoldFunctionAndIndexesQuery); err != nil {
		t.Fatalf("migration must not fail without extensions: %v", err)
	}

	var foldedName string
	if err := restricted.QueryRow(`SELECT kuranas_fold('Relatório')`).Scan(&foldedName); err != nil {
		t.Fatalf("kuranas_fold must exist without unaccent: %v", err)
	}
	if foldedName != "relatório" {
		t.Fatalf("fallback must only lowercase, got %q", foldedName)
	}
}

func restrictedConnectionString(role string, databaseName string) string {
	host := envOrDefault("TEST_DB_HOST", envOrDefault("DB_HOST", "127.0.0.1"))
	port := envOrDefault("TEST_DB_PORT", envOrDefault("DB_PORT", "5432"))
	return fmt.Sprintf("host=%s port=%s user=%s password=restricted dbname=%s sslmode=disable", host, port, role, databaseName)
}

func envOrDefault(key string, fallback string) string {
	if configuredValue := os.Getenv(key); configuredValue != "" {
		return configuredValue
	}
	return fallback
}
