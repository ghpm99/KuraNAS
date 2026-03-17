package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/analytics"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/search"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB            *database.DbContext
	Logger        logger.LoggerServiceInterface
	Tasks         *chan utils.Task
	Files         *FileContext
	Jobs          *JobsContext
	Diary         *DiaryContext
	Music         *MusicContext
	Video         *VideoContext
	Analytics     *AnalyticsContext
	Configuration *ConfigurationContext
	Search        *SearchContext
	Notifications *NotificationContext
	UpdateHandler *updater.Handler
	UpdateService *updater.Service
}

type FileContext struct {
	Handler              *files.Handler
	Service              files.ServiceInterface
	RecentFileService    files.RecentFileServiceInterface
	Repository           files.RepositoryInterface
	RecentFileRepository files.RecentFileRepositoryInterface
	MetadataRepository   files.MetadataRepositoryInterface
}

type JobsContext struct {
	Handler    *jobs.Handler
	Service    jobs.ServiceInterface
	Repository jobs.RepositoryInterface
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

type VideoContext struct {
	Handler    *video.Handler
	Service    video.ServiceInterface
	Repository video.RepositoryInterface
}

type AnalyticsContext struct {
	Handler    *analytics.Handler
	Service    analytics.ServiceInterface
	Repository analytics.RepositoryInterface
}

type ConfigurationContext struct {
	Handler    *configuration.Handler
	Service    configuration.ServiceInterface
	Repository configuration.RepositoryInterface
}

type SearchContext struct {
	Handler    *search.Handler
	Service    search.ServiceInterface
	Repository search.RepositoryInterface
}

type NotificationContext struct {
	Handler    *notifications.Handler
	Service    notifications.ServiceInterface
	Repository notifications.RepositoryInterface
}

func NewContext(db *sql.DB) *AppContext {

	dbContext := database.NewDbContext(db)

	loggerService := logger.NewLoggerService(logger.NewLoggerRepository(dbContext))
	jobsContext := newJobsContext(dbContext)
	fileContext := newFileContext(dbContext, loggerService, jobsContext.Repository)
	diaryContext := newDiaryContext(dbContext, loggerService)
	musicContext := newMusicContext(dbContext, loggerService)
	videoContext := newVideoContext(dbContext, loggerService)
	analyticsContext := newAnalyticsContext(dbContext)
	configurationContext := newConfigurationContext(dbContext, loggerService)
	searchContext := newSearchContext(dbContext)
	notificationContext := newNotificationContext(dbContext)
	updateService := updater.NewService()
	updateHandler := updater.NewHandler(updateService, loggerService)

	context := &AppContext{
		DB:            dbContext,
		Logger:        loggerService,
		Tasks:         &tasks,
		Files:         fileContext,
		Jobs:          jobsContext,
		Diary:         diaryContext,
		Music:         musicContext,
		Video:         videoContext,
		Analytics:     analyticsContext,
		Configuration: configurationContext,
		Search:        searchContext,
		Notifications: notificationContext,
		UpdateHandler: updateHandler,
		UpdateService: updateService,
	}
	return context
}

func newJobsContext(dbContext *database.DbContext) *JobsContext {
	repository := jobs.NewRepository(dbContext)
	service := jobs.NewService(repository)
	handler := jobs.NewHandler(service)
	return &JobsContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newFileContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface, jobsRepository jobs.RepositoryInterface) *FileContext {
	repository := files.NewRepository(dbContext)
	recentFileRepository := files.NewRecentFileRepository(dbContext)

	metadataRepository := files.NewMetadataRepository(dbContext)
	service := files.NewService(repository, metadataRepository, jobsRepository, tasks)
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

func newVideoContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface) *VideoContext {
	repository := video.NewRepository(dbContext)
	service := video.NewService(repository)
	handler := video.NewHandler(service, logger)
	return &VideoContext{
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

func newAnalyticsContext(dbContext *database.DbContext) *AnalyticsContext {
	repository := analytics.NewRepository(dbContext)
	service := analytics.NewService(repository)
	handler := analytics.NewHandler(service)
	return &AnalyticsContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newConfigurationContext(dbContext *database.DbContext, loggerService logger.LoggerServiceInterface) *ConfigurationContext {
	repository := configuration.NewRepository(dbContext)
	service := configuration.NewService(repository)
	handler := configuration.NewHandler(service, loggerService)
	_ = service.ApplyRuntimeSettings()

	return &ConfigurationContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newNotificationContext(dbContext *database.DbContext) *NotificationContext {
	repository := notifications.NewRepository(dbContext)
	service := notifications.NewService(repository)
	handler := notifications.NewHandler(service)

	return &NotificationContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newSearchContext(dbContext *database.DbContext) *SearchContext {
	repository := search.NewRepository(dbContext)
	service := search.NewService(repository)
	handler := search.NewHandler(service)

	return &SearchContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
