package files

import (
	"database/sql"
	"errors"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

// One small, optimized query per repository call: each method below answers
// exactly one question and embeds exactly one .sql (no optional-filter
// god-query — see docs/melhorias/08).

func scanFileRows(rows *sql.Rows) ([]FileModel, error) {
	items := []FileModel{}
	for rows.Next() {
		var file FileModel
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.Path,
			&file.ParentPath,
			&file.Format,
			&file.Size,
			&file.UpdatedAt,
			&file.CreatedAt,
			&file.LastInteraction,
			&file.LastBackup,
			&file.Type,
			&file.CheckSum,
			&file.DeletedAt,
			&file.Starred,
			&file.PhysicalPath,
		); err != nil {
			return nil, err
		}
		items = append(items, file)
	}
	return items, rows.Err()
}

// queryFilesPage runs a paginated file query that ends in LIMIT/OFFSET,
// fetching pageSize+1 rows so UpdatePagination can derive HasNext.
func (r *Repository) queryFilesPage(query string, page int, pageSize int, args ...any) (utils.PaginationResponse[FileModel], error) {
	response := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
	}

	queryArgs := append(args, pageSize+1, utils.CalculateOffset(page, pageSize))

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		items, scanErr := scanFileRows(rows)
		if scanErr != nil {
			return scanErr
		}
		response.Items = items
		return nil
	})
	if err != nil {
		return response, fmt.Errorf("failed to query files: %w", err)
	}

	response.UpdatePagination()
	return response, nil
}

// GetFileByID returns the row with the given id in any soft-delete state; the
// second return value reports whether it exists.
func (r *Repository) GetFileByID(id int) (FileModel, bool, error) {
	var file FileModel
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetFileByIDQuery, id)
		if err != nil {
			return err
		}
		defer rows.Close()

		items, scanErr := scanFileRows(rows)
		if scanErr != nil {
			return scanErr
		}
		if len(items) > 0 {
			file = items[0]
			found = true
		}
		return nil
	})
	if err != nil {
		return FileModel{}, false, fmt.Errorf("GetFileByID: %w", err)
	}
	return file, found, nil
}

// GetFilesByNameAndPath returns every row (active and soft-deleted) at an
// exact name+path, newest first, capped at limit.
func (r *Repository) GetFilesByNameAndPath(name string, path string, limit int) ([]FileModel, error) {
	if limit <= 0 {
		return nil, errors.New("GetFilesByNameAndPath: limit must be positive")
	}

	var items []FileModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetFilesByNameAndPathQuery, name, path, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		scanned, scanErr := scanFileRows(rows)
		if scanErr != nil {
			return scanErr
		}
		items = scanned
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetFilesByNameAndPath: %w", err)
	}
	return items, nil
}

// GetActiveChildrenByParentPath lists the active children of a directory,
// optionally narrowed to a category (starred / recently accessed).
func (r *Repository) GetActiveChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	query := queries.GetChildrenByParentPathQuery
	switch category {
	case StarredCategory:
		query = queries.GetStarredChildrenByParentPathQuery
	case RecentCategory:
		query = queries.GetRecentChildrenByParentPathQuery
	}
	return r.queryFilesPage(query, page, pageSize, parentPath)
}

// GetActiveFilesByPath returns the active row(s) at an exact path.
func (r *Repository) GetActiveFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetFilesByPathQuery, page, pageSize, path)
}

// GetActiveFiles lists all active files, paginated.
func (r *Repository) GetActiveFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetActiveFilesQuery, page, pageSize)
}

// GetFilesByPathPrefix walks a subtree (root row included) in any soft-delete
// state, paginated — the mark_deleted reconciliation feed.
func (r *Repository) GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetFilesByPathPrefixQuery, page, pageSize, prefix)
}
