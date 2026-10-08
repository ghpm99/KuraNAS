package files

import (
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

func (r *Repository) SearchActiveFilesByName(namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.SearchActiveFilesByNameQuery, page, pageSize, namePattern)
}

func (r *Repository) SearchActiveFilesByNameUnderPath(pathPrefix string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.SearchActiveFilesByNameUnderPathQuery, page, pageSize, pathPrefix, namePattern)
}

func (r *Repository) SearchActiveChildrenByName(parentPath string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.SearchActiveChildrenByNameQuery, page, pageSize, parentPath, namePattern)
}
