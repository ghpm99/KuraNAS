package files

import (
	"database/sql"
	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"
)

type RecentFileRepository struct {
	DbContext *sql.DB
}

func NewRecentFileRepository(db *sql.DB) *RecentFileRepository {
	return &RecentFileRepository{DbContext: db}
}

func (r *RecentFileRepository) Upsert(ip string, fileID int) error {
	_, err := r.DbContext.Exec(
		queries.UpsertRecentFileQuery,
		ip, fileID,
	)
	return err
}

func (r *RecentFileRepository) DeleteOld(ip string, keep int) error {
	_, err := r.DbContext.Exec(
		queries.DeleteOldRecentFilesQuery,
		ip, ip, keep,
	)
	return err
}

func (r *RecentFileRepository) GetRecentFiles(page int, pageSize int) ([]RecentFileModel, error) {
	rows, err := r.DbContext.Query(
		queries.GetRecentFilesQuery,
		pageSize+1,
		utils.CalculateOffset(page, pageSize),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentFileModel
	for rows.Next() {
		var rf RecentFileModel
		if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
			return nil, err
		}
		result = append(result, rf)
	}
	return result, nil
}

func (r *RecentFileRepository) Delete(ip string, fileID int) error {
	_, err := r.DbContext.Exec(
		queries.DeleteRecentFileQuery,
		ip, fileID,
	)
	return err
}

func (r *RecentFileRepository) GetByFileID(fileID int) ([]RecentFileModel, error) {
	rows, err := r.DbContext.Query(
		queries.GetRecentByFileIDQuery,
		fileID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RecentFileModel
	for rows.Next() {
		var rf RecentFileModel
		if err := rows.Scan(&rf.ID, &rf.IPAddress, &rf.FileID, &rf.AccessedAt); err != nil {
			return nil, err
		}
		result = append(result, rf)
	}
	return result, nil
}
