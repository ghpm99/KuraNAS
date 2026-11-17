package takeout

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/takeout"
)

type SidecarRepository struct {
	DbContext *database.DbContext
}

func NewSidecarRepository(dbContext *database.DbContext) *SidecarRepository {
	return &SidecarRepository{DbContext: dbContext}
}

func (r *SidecarRepository) SaveSidecars(sidecars []PhotoSidecar) error {
	err := r.DbContext.ExecTx(func(transaction *sql.Tx) error {
		for _, sidecar := range sidecars {
			if _, err := transaction.Exec(queries.UpsertTakeoutPhotoSidecarQuery, sidecar.Path, sql.NullString{String: sidecar.Description, Valid: sidecar.Description != ""}, sidecar.Latitude, sidecar.Longitude); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("SaveSidecars: %w", err)
	}
	return nil
}

func (r *SidecarRepository) FindSidecarByPath(path string) (*PhotoSidecar, error) {
	sidecar := PhotoSidecar{Path: path}
	var description sql.NullString
	var latitude, longitude sql.NullFloat64
	err := r.DbContext.GetDatabase().QueryRow(queries.SelectTakeoutPhotoSidecarByPathQuery, path).Scan(&description, &latitude, &longitude)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindSidecarByPath: %w", err)
	}
	sidecar.Description = description.String
	if latitude.Valid && longitude.Valid {
		sidecar.Latitude, sidecar.Longitude = &latitude.Float64, &longitude.Float64
	}
	return &sidecar, nil
}
