package app

import (
	"database/sql"
	"log"
	"nas-go/api/internal/api/v1/analytics"
	"nas-go/api/internal/api/v1/captures"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/search"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/providers/anthropic"
	"nas-go/api/pkg/ai/providers/openai"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB            *database.DbContext
	Logger        logger.LoggerServiceInterface
	AI            ai.ServiceInterface
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
	Captures      *CapturesContext
	Libraries     *LibrariesContext
	WatchFolders  *WatchFoldersContext
	Takeout       *TakeoutContext
	UpdateHandler *updater.Handler
	UpdateService *updater.Service
}

type CapturesContext struct {
	Handler    *captures.Handler
	Service    captures.ServiceInterface
	Repository captures.RepositoryInterface
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

type LibrariesContext struct {
	Handler    *libraries.Handler
	Service    libraries.ServiceInterface
	Repository libraries.RepositoryInterface
}

type WatchFoldersContext struct {
	Handler    *watchfolders.Handler
	Service    watchfolders.ServiceInterface
	Repository watchfolders.RepositoryInterface
}

type TakeoutContext struct {
	Handler *takeout.Handler
	Service takeout.ServiceInterface
}

func NewContext(db *sql.DB) *AppContext {

	dbContext := database.NewDbContext(db)

	loggerService := logger.NewLoggerService(logger.NewLoggerRepository(dbContext))
	aiService := newAIService()
	jobsContext := newJobsContext(dbContext)
	fileContext := newFileContext(dbContext, loggerService, jobsContext.Repository)
	diaryContext := newDiaryContext(dbContext, loggerService)
	musicContext := newMusicContext(dbContext, loggerService)
	videoContext := newVideoContext(dbContext, loggerService, aiService)
	analyticsContext := newAnalyticsContext(dbContext, aiService)
	configurationContext := newConfigurationContext(dbContext, loggerService)
	searchContext := newSearchContext(dbContext, aiService)
	notificationContext := newNotificationContext(dbContext)
	capturesContext := newCapturesContext(dbContext, loggerService, fileContext.Service, notificationContext.Service)
	librariesContext := newLibrariesContext(dbContext, loggerService)
	watchFoldersContext := newWatchFoldersContext(dbContext, loggerService)
	takeoutContext := newTakeoutContext(dbContext, loggerService, librariesContext.Service, jobsContext.Repository, notificationContext.Service)
	updateService := updater.NewService()
	updateHandler := updater.NewHandler(updateService, loggerService)

	context := &AppContext{
		DB:            dbContext,
		Logger:        loggerService,
		AI:            aiService,
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
		Captures:      capturesContext,
		Libraries:     librariesContext,
		WatchFolders:  watchFoldersContext,
		Takeout:       takeoutContext,
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

func newVideoContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface, aiService ai.ServiceInterface) *VideoContext {
	repository := video.NewRepository(dbContext)
	service := video.NewService(repository, aiService)
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

func newAnalyticsContext(dbContext *database.DbContext, aiService ai.ServiceInterface) *AnalyticsContext {
	repository := analytics.NewRepository(dbContext)
	service := analytics.NewService(repository, aiService)
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
	if dbContext != nil && dbContext.GetDatabase() != nil {
		_ = service.ApplyRuntimeSettings()
	}

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

func newSearchContext(dbContext *database.DbContext, aiService ai.ServiceInterface) *SearchContext {
	repository := search.NewRepository(dbContext)
	service := search.NewService(repository, aiService)
	handler := search.NewHandler(service)

	return &SearchContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newLibrariesContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
) *LibrariesContext {
	repository := libraries.NewRepository(dbContext)
	service := libraries.NewService(repository)
	handler := libraries.NewHandler(service, loggerService)

	return &LibrariesContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newWatchFoldersContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
) *WatchFoldersContext {
	repository := watchfolders.NewRepository(dbContext)
	service := watchfolders.NewService(repository)
	handler := watchfolders.NewHandler(service, loggerService)

	return &WatchFoldersContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newCapturesContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
	uploadJobDispatcher captures.UploadJobDispatcherInterface,
	notificationService notifications.ServiceInterface,
) *CapturesContext {
	repository := captures.NewRepository(dbContext)
	service := captures.NewService(repository, uploadJobDispatcher, notificationService)
	handler := captures.NewHandler(service, loggerService)
	return &CapturesContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newAIService() ai.ServiceInterface {
	cfg := ai.LoadConfig()
	if cfg.OpenAIAPIKey == "" && cfg.AnthropicAPIKey == "" {
		log.Println("AI service disabled: no API keys configured")
		return nil
	}

	router := ai.NewRouter()

	var primary ai.Provider
	var fallback ai.Provider

	if cfg.OpenAIAPIKey != "" {
		primary = openai.NewProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL, cfg.DefaultTimeout)
		log.Printf("AI provider registered: openai (%s)\n", cfg.OpenAIModel)
	}
	if cfg.AnthropicAPIKey != "" {
		provider := anthropic.NewProvider(cfg.AnthropicAPIKey, cfg.AnthropicModel, cfg.DefaultTimeout)
		if primary == nil {
			primary = provider
		} else {
			fallback = provider
		}
		log.Printf("AI provider registered: anthropic (%s)\n", cfg.AnthropicModel)
	}

	taskTypes := []ai.TaskType{
		ai.TaskClassification,
		ai.TaskExtraction,
		ai.TaskSummarization,
		ai.TaskGeneration,
		ai.TaskSimple,
		ai.TaskComplex,
	}

	for _, taskType := range taskTypes {
		if fallback != nil {
			router.RegisterWithFallback(taskType, primary, fallback)
		} else {
			router.Register(taskType, primary)
		}
	}

	log.Println("AI service enabled")
	return ai.NewService(router, cfg)
}

func newTakeoutContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
	libraryResolver takeout.LibraryResolverInterface,
	jobsRepository jobs.RepositoryInterface,
	notificationService notifications.ServiceInterface,
) *TakeoutContext {
	service := takeout.NewService(jobsRepository, libraryResolver, notificationService)
	handler := takeout.NewHandler(service, loggerService)
	return &TakeoutContext{
		Handler: handler,
		Service: service,
	}
}
