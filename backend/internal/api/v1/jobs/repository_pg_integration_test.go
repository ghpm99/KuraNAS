package jobs_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/internal/testutil"
)

func truncateWorkerTables(t *testing.T, ctx *database.DbContext) {
	t.Helper()
	if err := ctx.ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("TRUNCATE worker_step, worker_job RESTART IDENTITY CASCADE")
		return e
	}); err != nil {
		t.Fatalf("truncate worker tables: %v", err)
	}
}

// TestCreateStepWithoutPayload_Postgres reproduces the production failure where
// enqueuing the ai_playlist_cluster job died with
// "pq: sintaxe de entrada inválida para o tipo de dados json".
//
// That step carries no payload, so StepModel.Payload was a nil []byte. lib/pq
// encodes a nil/empty []byte as the empty string "" rather than SQL NULL, and
// the payload JSON column rejects "" as invalid JSON. The expected behaviour is
// that a payload-less step persists with payload (and depends_on) as NULL.
func TestCreateStepWithoutPayload_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_jobs_it")
	truncateWorkerTables(t, dbCtx)

	repo := jobs.NewRepository(dbCtx)

	var created jobs.StepModel
	err := dbCtx.ExecTx(func(tx *sql.Tx) error {
		job, jobErr := repo.CreateJob(tx, jobs.JobModel{
			Type:     "ai_playlist_cluster",
			Priority: "low",
			Scope:    []byte("{}"),
			Status:   "queued",
		})
		if jobErr != nil {
			return jobErr
		}

		step, stepErr := repo.CreateStep(tx, jobs.StepModel{
			JobID:       job.ID,
			Type:        "ai_playlist_cluster",
			Status:      "queued",
			MaxAttempts: 1,
			// DependsOn and Payload intentionally left as nil []byte — this is
			// exactly how enqueueAIPlaylistClusterJob plans the step.
		})
		if stepErr != nil {
			return stepErr
		}
		created = step
		return nil
	})
	if err != nil {
		t.Fatalf("CreateStep without payload must succeed, got: %v", err)
	}

	var payload sql.NullString
	var dependsOn sql.NullString
	readErr := dbCtx.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			"SELECT payload, depends_on FROM worker_step WHERE id = $1",
			created.ID,
		).Scan(&payload, &dependsOn)
	})
	if readErr != nil {
		t.Fatalf("read back step: %v", readErr)
	}

	if payload.Valid {
		t.Fatalf("expected payload to persist as NULL, got %q", payload.String)
	}
	if dependsOn.Valid {
		t.Fatalf("expected depends_on to persist as NULL, got %q", dependsOn.String)
	}
}
