package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB                   *sql.DB
	Logger               logger.LoggerServiceInterface
	Tasks                *chan utils.Task
	Files                *FileContext
	Diary                *DiaryContext
	ConfigurationHandler *configuration.Handler
}

type FileContext struct {
	Handler              *files.Handler
	Service              files.ServiceInterface
	RecentFileService    files.RecentFileServiceInterface
	Repository           files.RepositoryInterface
	RecentFileRepository files.RecentFileRepositoryInterface
	MetadataRepository   files.MetadataRepositoryInterface
}

type DiaryContext struct {
	Handler    *diary.Handler
	Service    diary.ServiceInterface
	Repository diary.RepositoryInterface
}

func NewContext(db *sql.DB) *AppContext {
	loggerService := logger.NewLoggerService(logger.NewLoggerRepository(db))
	fileContext := newFileContext(db, loggerService)
	diaryContext := newDiaryContext(db, loggerService)
	configurationHandler := configuration.NewHandler(loggerService)

	context := &AppContext{
		DB:                   db,
		Logger:               loggerService,
		Tasks:                &tasks,
		Files:                fileContext,
		Diary:                diaryContext,
		ConfigurationHandler: configurationHandler,
	}
	return context
}

func newFileContext(db *sql.DB, logger logger.LoggerServiceInterface) *FileContext {
	repository := files.NewRepository(db)
	recentFileRepository := files.NewRecentFileRepository(db)

	service := files.NewService(repository, tasks)
	recentFileService := files.NewRecentFileService(recentFileRepository)

	metadataRepository := files.NewMetadataRepository(db)

	handler := files.NewHandler(service, recentFileService, logger)
	return &FileContext{
		Handler:              handler,
		Service:              service,
		RecentFileService:    recentFileService,
		Repository:           repository,
		RecentFileRepository: recentFileRepository,
		MetadataRepository:   metadataRepository,
	}
}

func newDiaryContext(db *sql.DB, logger logger.LoggerServiceInterface) *DiaryContext {
	repository := diary.NewRepository(db)
	service := diary.NewService(repository, tasks)
	handler := diary.NewHandler(service, logger)
	return &DiaryContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
