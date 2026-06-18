package captures

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/captures"
	"nas-go/api/pkg/utils"
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

// rowScanner is satisfied by both *sql.Row and *sql.Rows so the column mapping
// lives in one place (scanCapture) for every read path.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanCapture(scanner rowScanner) (CaptureModel, error) {
	var m CaptureModel
	var (
		fileID, season, episode, releaseYear sql.NullInt64
		status                               string
		genres, cast, directors, rawMetadata []byte
	)

	err := scanner.Scan(
		&m.ID,
		&m.Name,
		&m.FileName,
		&m.FilePath,
		&m.MediaType,
		&m.MimeType,
		&m.Size,
		&m.EpisodeKey,
		&m.CreatedAt,
		&fileID,
		&status,
		&m.Title,
		&m.EpisodeTitle,
		&season,
		&episode,
		&m.Description,
		&releaseYear,
		&genres,
		&cast,
		&directors,
		&m.Studio,
		&m.ContentRating,
		&m.Platform,
		&m.SourceURL,
		&m.ThumbnailURL,
		&m.ContentType,
		&rawMetadata,
	)
	if err != nil {
		return m, err
	}

	m.Status = CaptureStatus(status)
	m.FileID = nullIntToPtr(fileID)
	m.Season = nullIntToPtr(season)
	m.Episode = nullIntToPtr(episode)
	m.ReleaseYear = nullIntToPtr(releaseYear)
	m.Genres = decodeStringArray(genres)
	m.Cast = decodeStringArray(cast)
	m.Directors = decodeStringArray(directors)
	m.RawMetadata = json.RawMessage(rawMetadata)

	return m, nil
}

func nullIntToPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

func ptrToNullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

// decodeStringArray turns a jsonb array column into []string, tolerating NULL,
// empty, or malformed payloads by returning nil rather than failing the scan.
func decodeStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func encodeStringArray(values []string) []byte {
	if len(values) == 0 {
		return []byte("[]")
	}
	data, err := json.Marshal(values)
	if err != nil {
		return []byte("[]")
	}
	return data
}

func rawMetadataOrEmpty(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func (r *Repository) CreateCapture(transaction *sql.Tx, capture CaptureModel) (CaptureModel, error) {
	status := capture.Status
	if status == "" {
		status = CaptureStatusUploaded
	}

	args := []any{
		capture.Name,
		capture.FileName,
		capture.FilePath,
		capture.MediaType,
		capture.MimeType,
		capture.Size,
		capture.EpisodeKey,
		capture.CreatedAt,
		string(status),
		rawMetadataOrEmpty(capture.RawMetadata),
	}

	var id int
	err := transaction.QueryRow(queries.InsertCaptureQuery, args...).Scan(&id)
	if err != nil {
		return capture, fmt.Errorf("CreateCapture: %w", err)
	}

	capture.ID = id
	capture.Status = status
	return capture, nil
}

func (r *Repository) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error) {
	paginationResponse := utils.PaginationResponse[CaptureModel]{
		Items: []CaptureModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.MediaType.HasValue,
		filter.MediaType.Value,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetCapturesQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			m, scanErr := scanCapture(rows)
			if scanErr != nil {
				return scanErr
			}
			paginationResponse.Items = append(paginationResponse.Items, m)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("GetCaptures: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetCaptureByID(id int) (CaptureModel, error) {
	var m CaptureModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		m, scanErr = scanCapture(tx.QueryRow(queries.GetCaptureByIDQuery, id))
		return scanErr
	})

	if err != nil {
		return m, fmt.Errorf("GetCaptureByID: %w", err)
	}

	return m, nil
}

// GetCaptureByEpisodeKey returns the most recent completed capture archived
// under the given episode_key. The boolean is false (with a nil error) when no
// capture matches, so the caller can distinguish "not archived yet" from a real
// query failure.
func (r *Repository) GetCaptureByEpisodeKey(episodeKey string) (CaptureModel, bool, error) {
	var m CaptureModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		m, scanErr = scanCapture(tx.QueryRow(queries.GetCaptureByEpisodeKeyQuery, episodeKey))
		return scanErr
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CaptureModel{}, false, nil
		}
		return m, false, fmt.Errorf("GetCaptureByEpisodeKey: %w", err)
	}

	return m, true, nil
}

// UpdateCapturePromotion writes the resolved file_id, final path and all the
// rich semantic columns parsed from the metadata, flipping the capture to its
// new status in the same statement.
func (r *Repository) UpdateCapturePromotion(transaction *sql.Tx, capture CaptureModel) error {
	args := []any{
		capture.ID,
		ptrToNullInt(capture.FileID),
		capture.FileName,
		capture.FilePath,
		string(capture.Status),
		capture.Title,
		capture.EpisodeTitle,
		ptrToNullInt(capture.Season),
		ptrToNullInt(capture.Episode),
		capture.Description,
		ptrToNullInt(capture.ReleaseYear),
		encodeStringArray(capture.Genres),
		encodeStringArray(capture.Cast),
		encodeStringArray(capture.Directors),
		capture.Studio,
		capture.ContentRating,
		capture.Platform,
		capture.SourceURL,
		capture.ThumbnailURL,
		capture.ContentType,
	}

	if _, err := transaction.Exec(queries.UpdateCapturePromotionQuery, args...); err != nil {
		return fmt.Errorf("UpdateCapturePromotion: %w", err)
	}
	return nil
}

// UpdateCaptureStatus sets a capture's lifecycle status and (optionally) clears
// or sets its file_id — used to flip to "failed" and detach the rolled-back
// home_file during a promotion failure.
func (r *Repository) UpdateCaptureStatus(transaction *sql.Tx, id int, status CaptureStatus, fileID *int) error {
	if _, err := transaction.Exec(queries.UpdateCaptureStatusQuery, id, string(status), ptrToNullInt(fileID)); err != nil {
		return fmt.Errorf("UpdateCaptureStatus: %w", err)
	}
	return nil
}

func (r *Repository) DeleteCapture(transaction *sql.Tx, id int) error {
	result, err := transaction.Exec(queries.DeleteCaptureQuery, id)
	if err != nil {
		return fmt.Errorf("DeleteCapture: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteCapture: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DeleteCapture: capture not found")
	}

	return nil
}
