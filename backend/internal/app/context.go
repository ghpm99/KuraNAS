package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
)

type AppContext struct {
	DB    *sql.DB
	Files *FileContext
}

type FileContext struct {
	Handler    *files.Handler
	Service    *files.Service
	Repository *files.Repository
}

func NewContext(db *sql.DB) *AppContext {
	fileContext := newFileContext(db)
	context := &AppContext{
		DB:    db,
		Files: fileContext,
	}
	return context
}

func newFileContext(db *sql.DB) *FileContext {
	repository := files.NewRepository(db)
	service := files.NewService(repository)
	handler := files.NewHandler(service)
	return &FileContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
