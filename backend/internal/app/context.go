package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB                   *database.DbContext
	Logger               logger.LoggerServiceInterface
	Tasks                *chan utils.Task
	Files                *FileContext
	Diary                *DiaryContext
	Music                *MusicContext
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

type MusicContext struct {
	Handler    *music.Handler
	Service    music.ServiceInterface
	Repository music.RepositoryInterface
}

func NewContext(db *sql.DB) *AppContext {

	dbContext := database.NewDbContext(db)

	loggerService := logger.NewLoggerService(logger.NewLoggerRepository(dbContext))
	fileContext := newFileContext(dbContext, loggerService)
	diaryContext := newDiaryContext(dbContext, loggerService)
	musicContext := newMusicContext(dbContext, loggerService)
	configurationHandler := configuration.NewHandler(loggerService)

	context := &AppContext{
		DB:                   dbContext,
		Logger:               loggerService,
		Tasks:                &tasks,
		Files:                fileContext,
		Diary:                diaryContext,
		Music:                musicContext,
		ConfigurationHandler: configurationHandler,
	}
	return context
}

func newFileContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface) *FileContext {
	repository := files.NewRepository(dbContext)
	recentFileRepository := files.NewRecentFileRepository(dbContext)

	metadataRepository := files.NewMetadataRepository(dbContext)
	service := files.NewService(repository, metadataRepository, tasks)
	recentFileService := files.NewRecentFileService(recentFileRepository)

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

func newMusicContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface) *MusicContext {
	repository := music.NewRepository(dbContext)
	service := music.NewService(repository)
	handler := music.NewHandler(service, logger)
	return &MusicContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newDiaryContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface) *DiaryContext {
	repository := diary.NewRepository(dbContext)
	service := diary.NewService(repository, tasks)
	handler := diary.NewHandler(service, logger)
	return &DiaryContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
