package tiering

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/tiering"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{DbContext: db}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) GetSettingsDocument() (string, bool, error) {
	var document string
	found := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(queries.GetTieringSettingsQuery).Scan(&document)
		if errors.Is(scanErr, sql.ErrNoRows) {
			found = false
			return nil
		}
		return scanErr
	})

	if err != nil {
		return "", false, fmt.Errorf("GetSettingsDocument: %w", err)
	}
	return document, found, nil
}

func (r *Repository) UpsertSettingsDocument(document string) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(queries.UpsertTieringSettingsQuery, document)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("UpsertSettingsDocument: %w", err)
	}
	return nil
}

func (r *Repository) ListDemotionCandidates(minSizeBytes int64, idleBefore time.Time) ([]CandidateModel, error) {
	var candidates []CandidateModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(queries.ListDemotionCandidatesQuery, minSizeBytes, idleBefore)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var candidate CandidateModel
			if scanErr := rows.Scan(&candidate.FileID, &candidate.LogicalPath, &candidate.Size); scanErr != nil {
				return scanErr
			}
			candidates = append(candidates, candidate)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("ListDemotionCandidates: %w", err)
	}
	return candidates, nil
}

func (r *Repository) ListPromotionCandidates(usedAfter time.Time) ([]CandidateModel, error) {
	var candidates []CandidateModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(queries.ListPromotionCandidatesQuery, usedAfter)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var candidate CandidateModel
			if scanErr := rows.Scan(&candidate.FileID, &candidate.LogicalPath, &candidate.PhysicalPath, &candidate.Size); scanErr != nil {
				return scanErr
			}
			candidates = append(candidates, candidate)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("ListPromotionCandidates: %w", err)
	}
	return candidates, nil
}

func (r *Repository) SetPhysicalPath(fileID int, physicalPath string) error {
	value := sql.NullString{String: physicalPath, Valid: physicalPath != ""}

	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(queries.SetPhysicalPathQuery, fileID, value)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("SetPhysicalPath: %w", err)
	}
	return nil
}

func (r *Repository) GetLastRun() (LastRunModel, bool, error) {
	var run LastRunModel
	found := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var startedAt, endedAt sql.NullTime
		scanErr := tx.QueryRow(queries.GetLastTieringJobQuery).Scan(
			&run.JobID, &run.Status, &run.CreatedAt, &startedAt, &endedAt, &run.LastError,
		)
		if errors.Is(scanErr, sql.ErrNoRows) {
			found = false
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		if startedAt.Valid {
			run.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			run.EndedAt = &endedAt.Time
		}
		return nil
	})

	if err != nil {
		return LastRunModel{}, false, fmt.Errorf("GetLastRun: %w", err)
	}
	return run, found, nil
}

func (r *Repository) GetTierCounts() (TierCountsModel, error) {
	var counts TierCountsModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetTierCountsQuery).Scan(
			&counts.HotFiles, &counts.HotBytes, &counts.ColdFiles, &counts.ColdBytes,
		)
	})

	if err != nil {
		return TierCountsModel{}, fmt.Errorf("GetTierCounts: %w", err)
	}
	return counts, nil
}
