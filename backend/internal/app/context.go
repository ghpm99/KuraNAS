package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB     *sql.DB
	Logger logger.LoggerServiceInterface
	Tasks  *chan utils.Task
	Files  *FileContext
	Diary  *DiaryContext
}

type FileContext struct {
	Handler    *files.Handler
	Service    files.ServiceInterface
	Repository files.RepositoryInterface
}

type DiaryContext struct {
	Handler    *diary.Handler
	Service    diary.ServiceInterface
	Repository diary.RepositoryInterface
}

func NewContext(db *sql.DB) *AppContext {
	LoggerService := logger.NewLoggerService(logger.NewLoggerRepository(db))
	fileContext := newFileContext(db, LoggerService)
	diaryContext := newDiaryContext(db)
	context := &AppContext{
		DB:     db,
		Logger: LoggerService,
		Tasks:  &tasks,
		Files:  fileContext,
		Diary:  diaryContext,
	}
	return context
}

func newFileContext(db *sql.DB, logger logger.LoggerServiceInterface) *FileContext {
	repository := files.NewRepository(db)
	service := files.NewService(repository, tasks, logger)
	handler := files.NewHandler(service)
	return &FileContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newDiaryContext(db *sql.DB) *DiaryContext {
	repository := diary.NewRepository(db)
	service := diary.NewService(repository, tasks)
	handler := diary.NewHandler(service)
	return &DiaryContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
