package analytics

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

type libraryMediaFixture struct {
	name           string
	hasMetadata    bool
	activeStepType string
	terminalStatus string
	isDeleted      bool
	metadataTable  string
}

func newLibraryMediaTypeRepository(t *testing.T) *Repository {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_analytics_library_it"))
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("TRUNCATE home_file, worker_job RESTART IDENTITY CASCADE")
		return execErr
	})
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
	return repo
}

func insertLibraryMediaFile(t *testing.T, repo *Repository, fixture libraryMediaFixture) {
	t.Helper()
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		var fileID int
		err := tx.QueryRow(
			`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, deleted_at)
			 VALUES ($1, $2, '/lib', $3, 10, NOW(), NOW(), 2, CASE WHEN $4 THEN NOW() END) RETURNING id`,
			fixture.name, "/lib/"+fixture.name, formatOfFileName(fixture.name), fixture.isDeleted,
		).Scan(&fileID)
		if err != nil {
			return err
		}
		if fixture.hasMetadata {
			if _, err := tx.Exec("INSERT INTO "+fixture.metadataTable+" (file_id, path) VALUES ($1, $2)", fileID, "/lib/"+fixture.name); err != nil {
				return err
			}
		}
		var jobID int
		if err := tx.QueryRow(`INSERT INTO worker_job (type, priority, scope, status) VALUES ('startup_scan', 'normal', NULL, 'completed') RETURNING id`).Scan(&jobID); err != nil {
			return err
		}
		if fixture.terminalStatus != "" {
			if _, err := tx.Exec(
				`INSERT INTO worker_step (job_id, type, status, payload, created_at, ended_at)
				 VALUES ($1, 'metadata', $2, $3, NOW(), NOW())`,
				jobID, fixture.terminalStatus, filePayload(fileID),
			); err != nil {
				return err
			}
		}
		if fixture.activeStepType != "" {
			if _, err := tx.Exec(
				`INSERT INTO worker_step (job_id, type, status, payload, created_at)
				 VALUES ($1, 'metadata', $2, $3, NOW())`,
				jobID, fixture.activeStepType, filePayload(fileID),
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert fixture %s: %v", fixture.name, err)
	}
}

func formatOfFileName(name string) string {
	for index := len(name) - 1; index >= 0; index-- {
		if name[index] == '.' {
			return name[index:]
		}
	}
	return ""
}

func libraryMediaTypesByName(t *testing.T, repo *Repository) map[string]LibraryMediaTypeModel {
	t.Helper()
	summary, err := repo.GetLibrarySummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mediaTypesByName := map[string]LibraryMediaTypeModel{}
	for _, mediaType := range summary.ByMediaType {
		mediaTypesByName[mediaType.MediaType] = mediaType
	}
	return mediaTypesByName
}

func TestPostgres_LibraryMediaTypeCountsTotalMetadataPendingAndFailed(t *testing.T) {
	repo := newLibraryMediaTypeRepository(t)
	fixtures := []libraryMediaFixture{
		{name: "a1.mp3", hasMetadata: true, metadataTable: "audio_metadata", terminalStatus: "completed"},
		{name: "a2.mp3"},
		{name: "a3.mp3", terminalStatus: "failed"},
		{name: "a4.mp3", terminalStatus: "failed", activeStepType: "queued"},
		{name: "a5.mp3", hasMetadata: true, metadataTable: "audio_metadata", isDeleted: true},
		{name: "v1.mp4", hasMetadata: true, metadataTable: "video_metadata"},
		{name: "v2.mp4", activeStepType: "running"},
		{name: "i1.jpg", hasMetadata: true, metadataTable: "image_metadata"},
		{name: "i2.jpg", terminalStatus: "failed"},
		{name: "i3.jpg", terminalStatus: "completed"},
		{name: "doc.pdf"},
	}
	for _, fixture := range fixtures {
		insertLibraryMediaFile(t, repo, fixture)
	}

	mediaTypesByName := libraryMediaTypesByName(t, repo)

	expected := map[string]LibraryMediaTypeModel{
		"audio": {MediaType: "audio", Total: 4, WithMetadata: 1, Pending: 2, Failed: 1},
		"video": {MediaType: "video", Total: 2, WithMetadata: 1, Pending: 1, Failed: 0},
		"image": {MediaType: "image", Total: 3, WithMetadata: 1, Pending: 1, Failed: 1},
	}
	for mediaTypeName, expectedModel := range expected {
		if mediaTypesByName[mediaTypeName] != expectedModel {
			t.Fatalf("%s: expected %+v, got %+v", mediaTypeName, expectedModel, mediaTypesByName[mediaTypeName])
		}
	}
	if len(mediaTypesByName) != len(expected) {
		t.Fatalf("unexpected media types: %v", mediaTypesByName)
	}
}

func TestPostgres_LibraryMediaTypeListsEveryTypeOnEmptyLibrary(t *testing.T) {
	repo := newLibraryMediaTypeRepository(t)

	summary, err := repo.GetLibrarySummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.ByMediaType) != 3 || summary.ByMediaType[0].MediaType != "audio" || summary.ByMediaType[2].Total != 0 {
		t.Fatalf("unexpected media types: %+v", summary.ByMediaType)
	}
}
