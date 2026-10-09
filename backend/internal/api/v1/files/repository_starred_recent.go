package files

import (
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

func (r *Repository) GetStarredFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetStarredFilesQuery, page, pageSize)
}

func (r *Repository) GetRecentlyAccessedFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetRecentlyAccessedFilesQuery, page, pageSize)
}
