package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB    *sql.DB
	Tasks *chan utils.Task
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
		Tasks: &tasks,
		Files: fileContext,
	}
	return context
}

func newFileContext(db *sql.DB) *FileContext {
	repository := files.NewRepository(db)
	service := files.NewService(repository)
	handler := files.NewHandler(service, tasks)
	return &FileContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
