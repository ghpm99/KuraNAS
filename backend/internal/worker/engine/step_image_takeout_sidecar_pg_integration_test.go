package engine

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	imagedom "nas-go/api/internal/api/v1/image"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type storedImageMetadata struct {
	description sql.NullString
	latitude    sql.NullFloat64
	longitude   sql.NullFloat64
}

func stubImageScript(t *testing.T, exifMetadata imagedom.MetadataModel) {
	t.Helper()
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		exifMetadata.Path = filePath
		exifMetadata.Format = "jpg"
		exifMetadata.Width, exifMetadata.Height = 10, 10
		encoded, err := json.Marshal(exifMetadata)
		return string(encoded), err
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })
}

func runMetadataStepForImage(t *testing.T, fixture reindexFixture, imageName string, sidecar takeout.PhotoSidecar) storedImageMetadata {
	t.Helper()
	imagePath := filepath.Join(fixture.scopedFolder, imageName)
	writeFixtureFile(t, imagePath)
	insertHomeFile(t, fixture.dbContext, imageName, imagePath, fixture.scopedFolder)

	sidecarRepository := takeout.NewSidecarRepository(fixture.dbContext)
	sidecar.Path = imagePath
	if err := sidecarRepository.SaveSidecars([]takeout.PhotoSidecar{sidecar}); err != nil {
		t.Fatalf("save sidecar: %v", err)
	}
	fixture.workerContext.ImageRepository = imagedom.NewRepository(fixture.dbContext)
	fixture.workerContext.TakeoutSidecars = sidecarRepository

	persisted, err := fixture.workerContext.FilesService.GetFileByNameAndPath(imageName, imagePath)
	if err != nil {
		t.Fatalf("load persisted file: %v", err)
	}
	payload, _ := marshalPayload(StepFilePayload{File: &persisted})
	if err := executeMetadataStep(fixture.workerContext, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMetadataStep: %v", err)
	}
	return loadStoredImageMetadata(t, fixture.dbContext, imagePath)
}

func loadStoredImageMetadata(t *testing.T, dbContext *database.DbContext, imagePath string) storedImageMetadata {
	t.Helper()
	var stored storedImageMetadata
	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT image_description, gps_latitude, gps_longitude FROM image_metadata WHERE path = $1`, imagePath).
			Scan(&stored.description, &stored.latitude, &stored.longitude)
	})
	if err != nil {
		t.Fatalf("read image_metadata for %q: %v", imagePath, err)
	}
	return stored
}

func sidecarWithLocation(description string) takeout.PhotoSidecar {
	latitude, longitude := -23.55, -46.63
	return takeout.PhotoSidecar{Description: description, Latitude: &latitude, Longitude: &longitude}
}

func TestMetadataStep_AppliesTakeoutSidecarWhenExifIsEmpty_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	stubImageScript(t, imagedom.MetadataModel{})

	stored := runMetadataStepForImage(t, fixture, "no-exif.jpg", sidecarWithLocation("Praia de Santos"))

	if stored.description.String != "Praia de Santos" {
		t.Fatalf("expected the sidecar description, got %+v", stored.description)
	}
	if stored.latitude.Float64 != -23.55 || stored.longitude.Float64 != -46.63 {
		t.Fatalf("expected the sidecar location, got %+v %+v", stored.latitude, stored.longitude)
	}
}

func TestMetadataStep_KeepsExifDescriptionAndLocationOverSidecar_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	stubImageScript(t, imagedom.MetadataModel{
		ImageDescription: "Descrição do EXIF",
		GPSLatitude:      10.5,
		GPSLongitude:     20.5,
	})

	stored := runMetadataStepForImage(t, fixture, "with-exif.jpg", sidecarWithLocation("Descrição do Takeout"))

	if stored.description.String != "Descrição do EXIF" {
		t.Fatalf("EXIF description must win, got %+v", stored.description)
	}
	if stored.latitude.Float64 != 10.5 || stored.longitude.Float64 != 20.5 {
		t.Fatalf("EXIF location must win, got %+v %+v", stored.latitude, stored.longitude)
	}
}

func TestMetadataStep_FillsOnlyTheFieldsTheExifLeftEmpty_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	stubImageScript(t, imagedom.MetadataModel{ImageDescription: "Só descrição no EXIF"})

	stored := runMetadataStepForImage(t, fixture, "partial-exif.jpg", sidecarWithLocation("Descrição do Takeout"))

	if stored.description.String != "Só descrição no EXIF" {
		t.Fatalf("EXIF description must win, got %+v", stored.description)
	}
	if stored.latitude.Float64 != -23.55 || stored.longitude.Float64 != -46.63 {
		t.Fatalf("missing EXIF location must come from the sidecar, got %+v %+v", stored.latitude, stored.longitude)
	}
}

func TestFillImageMetadataGapsFromSidecar_ReportsWhetherAnythingChanged(t *testing.T) {
	filled := imagedom.MetadataModel{}
	if !fillImageMetadataGapsFromSidecar(&filled, sidecarWithLocation("x")) || filled.ImageDescription != "x" || filled.GPSLatitude != -23.55 {
		t.Fatalf("expected both gaps filled, got %+v", filled)
	}

	complete := imagedom.MetadataModel{ImageDescription: "exif", GPSLatitude: 1, GPSLongitude: 2}
	if fillImageMetadataGapsFromSidecar(&complete, sidecarWithLocation("x")) || complete.ImageDescription != "exif" || complete.GPSLatitude != 1 {
		t.Fatalf("complete EXIF must stay untouched, got %+v", complete)
	}

}
