package libraries

import (
	"database/sql"
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAll() ([]LibraryModel, error)
	GetByCategory(category LibraryCategory) (LibraryModel, error)
	Upsert(tx *sql.Tx, model LibraryModel) (LibraryModel, error)
}

type ServiceInterface interface {
	GetLibraries() ([]LibraryDto, error)
	GetLibraryByCategory(category LibraryCategory) (LibraryDto, error)
	UpdateLibrary(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error)
	ResolveLibraries() error
}
