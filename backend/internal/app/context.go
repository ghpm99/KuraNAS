package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"nas-go/api/internal/api/v1/accesscontrol"
	"nas-go/api/internal/api/v1/aiproviders"
	"nas-go/api/internal/api/v1/analytics"
	"nas-go/api/internal/api/v1/assistant"
	"nas-go/api/internal/api/v1/backup"
	"nas-go/api/internal/api/v1/captures"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/distribution"
	"nas-go/api/internal/api/v1/email"
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/notifications"
	ollamamgmt "nas-go/api/internal/api/v1/ollama"
	"nas-go/api/internal/api/v1/search"
	"nas-go/api/internal/api/v1/storageroots"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/internal/api/v1/tiering"
	"nas-go/api/internal/api/v1/trash"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/agent"
	"nas-go/api/pkg/ai/providers/anthropic"
	"nas-go/api/pkg/ai/providers/ollama"
	"nas-go/api/pkg/ai/providers/openai"
	"nas-go/api/pkg/crypto"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

var tasks = make(chan utils.Task, 100)

type AppContext struct {
	DB            *database.DbContext
	Logger        logger.LoggerServiceInterface
	AccessControl *AccessControlContext
	AI            ai.ServiceInterface
	Assistant     *AssistantContext
	AIProviders   *AIProvidersContext
	Ollama        *OllamaContext
	Tasks         *chan utils.Task
	Files         *FileContext
	Image         *ImageContext
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
	StorageRoots  *StorageRootsContext
	Email         *EmailContext
	Takeout       *TakeoutContext
	Trash         *TrashContext
	Backup        *BackupContext
	Tiering       *TieringContext
	Distribution  *DistributionContext
	UpdateHandler *updater.Handler
	UpdateService *updater.Service
}

type AccessControlContext struct {
	Handler    *accesscontrol.Handler
	Service    accesscontrol.ServiceInterface
	Repository accesscontrol.RepositoryInterface
}

type TrashContext struct {
	Handler    *trash.Handler
	Service    trash.ServiceInterface
	Repository trash.RepositoryInterface
}

type BackupContext struct {
	Handler    *backup.Handler
	Service    backup.ServiceInterface
	Repository backup.RepositoryInterface
}

type TieringContext struct {
	Handler    *tiering.Handler
	Service    tiering.ServiceInterface
	Repository tiering.RepositoryInterface
}

type StorageRootsContext struct {
	Handler    *storageroots.Handler
	Service    storageroots.ServiceInterface
	Repository storageroots.RepositoryInterface
}

type EmailContext struct {
	Handler    *email.Handler
	Service    email.ServiceInterface
	Repository email.RepositoryInterface
}

type DistributionContext struct {
	Handler    *distribution.Handler
	Service    distribution.ServiceInterface
	Repository distribution.RepositoryInterface
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
}

type ImageContext struct {
	Handler    *imagedom.Handler
	Service    imagedom.ServiceInterface
	Repository imagedom.RepositoryInterface
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
	Handler                 *music.Handler
	Service                 music.ServiceInterface
	Repository              music.RepositoryInterface
	AudioMetadataRepository music.AudioMetadataRepositoryInterface
}

type VideoContext struct {
	Handler            *video.Handler
	Service            video.ServiceInterface
	Repository         video.RepositoryInterface
	MetadataRepository video.VideoMetadataRepositoryInterface
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

type AIProvidersContext struct {
	Handler    *aiproviders.Handler
	Service    aiproviders.ServiceInterface
	Repository aiproviders.RepositoryInterface
}

type OllamaContext struct {
	Handler *ollamamgmt.Handler
	Service ollamamgmt.ServiceInterface
	Daemon  *ollamamgmt.DaemonManager
}

type AssistantContext struct {
	Handler    *assistant.Handler
	Service    assistant.ServiceInterface
	Repository assistant.RepositoryInterface
}

func NewContext(db *sql.DB) *AppContext {

	dbContext := database.NewDbContext(db)

	loggerService := logger.NewLoggerService(logger.NewLoggerRepository(dbContext))
	aiService, aiProvidersContext := newAIStack(dbContext)
	jobsContext := newJobsContext(dbContext)
	ollamaContext := newOllamaContext(aiProvidersContext.Service, jobsContext.Repository)
	fileContext := newFileContext(dbContext, loggerService, jobsContext.Repository)
	imageContext := newImageContext(dbContext, loggerService)
	diaryContext := newDiaryContext(dbContext, loggerService)
	musicContext := newMusicContext(dbContext, loggerService, aiService, fileContext.Service)
	videoContext := newVideoContext(dbContext, loggerService, aiService, fileContext.Service)
	analyticsContext := newAnalyticsContext(dbContext, aiService)
	configurationContext := newConfigurationContext(dbContext, loggerService)
	searchContext := newSearchContext(dbContext, aiService)
	notificationContext := newNotificationContext(dbContext)
	capturesContext := newCapturesContext(dbContext, loggerService, fileContext.Service, notificationContext.Service)
	librariesContext := newLibrariesContext(dbContext, loggerService)
	watchFoldersContext := newWatchFoldersContext(dbContext, loggerService)
	accessControlContext := newAccessControlContext(dbContext, loggerService)
	trashContext := newTrashContext(dbContext, loggerService, fileContext.Service)
	// Delete-to-trash: files cannot import trash (trash already leans on
	// files), so the bin arrives by interface after both contexts exist.
	fileContext.Service.SetTrashBin(trashContext.Service)
	storageRootsContext := newStorageRootsContext(dbContext, loggerService, fileContext.Service)
	// Load the registered roots (seeding ENTRY_POINT on first boot) before
	// workers and watchers start consuming the registry.
	if err := storageRootsContext.Service.ReloadRegistry(); err != nil {
		log.Printf("storageroots: registry load failed (falling back to ENTRY_POINT): %v", err)
	}
	backupContext := newBackupContext(dbContext)
	tieringContext := newTieringContext(dbContext)
	emailContext := newEmailContext(dbContext, jobsContext.Repository)
	takeoutContext := newTakeoutContext(dbContext, loggerService, librariesContext.Service, jobsContext.Repository, notificationContext.Service)
	distributionContext := newDistributionContext()
	updateService := updater.NewService()
	updateHandler := updater.NewHandler(updateService, loggerService)
	assistantAgent := buildAssistantAgent(aiService, searchContext.Service)
	assistantContext := newAssistantContext(dbContext, aiService, assistantAgent)

	context := &AppContext{
		DB:            dbContext,
		Logger:        loggerService,
		AccessControl: accessControlContext,
		AI:            aiService,
		Assistant:     assistantContext,
		AIProviders:   aiProvidersContext,
		Ollama:        ollamaContext,
		Tasks:         &tasks,
		Files:         fileContext,
		Image:         imageContext,
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
		StorageRoots:  storageRootsContext,
		Email:         emailContext,
		Takeout:       takeoutContext,
		Trash:         trashContext,
		Backup:        backupContext,
		Tiering:       tieringContext,
		Distribution:  distributionContext,
		UpdateHandler: updateHandler,
		UpdateService: updateService,
	}
	return context
}

// newAssistantContext builds the conversational chat module: the hot-swappable
// AI service, a repository persisting conversations/history, and the tool-calling
// agent.
func newAssistantContext(dbContext *database.DbContext, aiService ai.ServiceInterface, assistantAgent assistant.AgentInterface) *AssistantContext {
	repository := assistant.NewRepository(dbContext)
	service := assistant.NewService(aiService, repository, assistantAgent)
	handler := assistant.NewHandler(service)
	return &AssistantContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

// buildAssistantAgent wires the tool registry. Tool handlers are constructed here
// (the composition root) so the generic agent package stays free of feature
// dependencies. Today it exposes one read-only tool: file search.
func buildAssistantAgent(aiService ai.ServiceInterface, searchService search.ServiceInterface) assistant.AgentInterface {
	registry := agent.NewRegistry()
	registry.Register(buildSearchTool(searchService))
	return agent.NewAgent(aiService, registry)
}

func buildSearchTool(searchService search.ServiceInterface) agent.Tool {
	return agent.Tool{
		Name:        "buscar_arquivos",
		Description: "Busca arquivos, pastas, músicas, vídeos e imagens no NAS a partir de um termo de busca.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Termo de busca, ex.: nome do arquivo ou assunto"}},"required":["query"]}`),
		Keywords: []string{
			"arquivo", "arquivos", "busca", "buscar", "procur", "acha", "achar", "encontr",
			"pasta", "foto", "fotos", "imagem", "imagens", "pdf", "documento", "documentos",
			"música", "musica", "músicas", "musicas", "vídeo", "video", "vídeos", "videos",
		},
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			var parsed struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(args, &parsed); err != nil {
				return "", fmt.Errorf("argumentos inválidos: %w", err)
			}
			query := strings.TrimSpace(parsed.Query)
			if query == "" {
				return "Nenhum termo de busca informado.", nil
			}
			result, err := searchService.SearchGlobal(query, 5)
			if err != nil {
				return "", err
			}
			return formatSearchResults(query, result), nil
		},
	}
}

func formatSearchResults(query string, result search.GlobalSearchResponseDto) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Resultados para %q:\n", query)
	total := 0
	for _, f := range result.Files {
		fmt.Fprintf(&b, "- Arquivo: %s (%s)\n", f.Name, f.Path)
		total++
	}
	for _, f := range result.Folders {
		fmt.Fprintf(&b, "- Pasta: %s (%s)\n", f.Name, f.Path)
		total++
	}
	for _, v := range result.Videos {
		fmt.Fprintf(&b, "- Vídeo: %s (%s)\n", v.Name, v.Path)
		total++
	}
	for _, img := range result.Images {
		fmt.Fprintf(&b, "- Imagem: %s (%s)\n", img.Name, img.Path)
		total++
	}
	for _, a := range result.Artists {
		fmt.Fprintf(&b, "- Artista: %s (%d faixas)\n", a.Artist, a.TrackCount)
		total++
	}
	if total == 0 {
		return fmt.Sprintf("Nenhum resultado encontrado para %q.", query)
	}
	return b.String()
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

	service := files.NewService(repository, jobsRepository, tasks)
	recentFileService := files.NewRecentFileService(recentFileRepository)

	handler := files.NewHandler(service, recentFileService, logger)
	return &FileContext{
		Handler:              handler,
		Service:              service,
		RecentFileService:    recentFileService,
		Repository:           repository,
		RecentFileRepository: recentFileRepository,
	}
}

func newImageContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface) *ImageContext {
	repository := imagedom.NewRepository(dbContext)
	service := imagedom.NewService(repository)
	handler := imagedom.NewHandler(service, logger)
	return &ImageContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newMusicContext(dbContext *database.DbContext, loggerSvc logger.LoggerServiceInterface, aiService ai.ServiceInterface, filesService files.ServiceInterface) *MusicContext {
	repository := music.NewRepository(dbContext)
	audioMetadataRepository := music.NewAudioMetadataRepository(dbContext)
	service := music.NewService(repository, aiService)
	handler := music.NewHandler(service, filesService, loggerSvc)
	return &MusicContext{
		Handler:                 handler,
		Service:                 service,
		Repository:              repository,
		AudioMetadataRepository: audioMetadataRepository,
	}
}

func newVideoContext(dbContext *database.DbContext, logger logger.LoggerServiceInterface, aiService ai.ServiceInterface, filesService files.ServiceInterface) *VideoContext {
	repository := video.NewRepository(dbContext)
	metadataRepository := video.NewVideoMetadataRepository(dbContext)
	service := video.NewService(repository, aiService)
	handler := video.NewHandler(service, filesService, logger)
	return &VideoContext{
		Handler:            handler,
		Service:            service,
		Repository:         repository,
		MetadataRepository: metadataRepository,
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

func newStorageRootsContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
	filesService files.ServiceInterface,
) *StorageRootsContext {
	repository := storageroots.NewRepository(dbContext)
	service := storageroots.NewService(repository, filesService)
	handler := storageroots.NewHandler(service, loggerService)

	return &StorageRootsContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newTrashContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
	filesService files.ServiceInterface,
) *TrashContext {
	repository := trash.NewRepository(dbContext)
	service := trash.NewService(repository, filesService)
	handler := trash.NewHandler(service, loggerService)

	return &TrashContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newBackupContext(dbContext *database.DbContext) *BackupContext {
	repository := backup.NewRepository(dbContext)
	service := backup.NewService(repository)
	handler := backup.NewHandler(service)

	return &BackupContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newTieringContext(dbContext *database.DbContext) *TieringContext {
	repository := tiering.NewRepository(dbContext)
	service := tiering.NewService(repository)
	handler := tiering.NewHandler(service)

	return &TieringContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

// newEmailContext wires the e-mail accounts module. Without a valid
// EMAIL_TOKEN_KEY the whole feature refuses to turn on (returns nil; the
// routes then answer EMAIL_FEATURE_DISABLED_NO_KEY) — tokens are never stored
// unencrypted.
func newEmailContext(dbContext *database.DbContext, jobsRepository jobs.RepositoryInterface) *EmailContext {
	key := config.AppConfig.EmailTokenKey
	if key == "" {
		log.Println("email: EMAIL_TOKEN_KEY not configured; e-mail integration is off")
		return nil
	}

	cipher, err := crypto.NewAESGCM(key)
	if err != nil {
		log.Printf("email: invalid EMAIL_TOKEN_KEY (%v); e-mail integration is off", err)
		return nil
	}

	repository := email.NewRepository(dbContext)
	service := email.NewService(repository, cipher, email.Config{
		GoogleClientID:        config.AppConfig.EmailGoogleClientID,
		GoogleClientSecret:    config.AppConfig.EmailGoogleClientSecret,
		MicrosoftClientID:     config.AppConfig.EmailMSClientID,
		RetentionDays:         config.AppConfig.EmailRetentionDays,
		MaxMessagesPerAccount: config.AppConfig.EmailMaxMessagesPerAccount,
	})
	// The jobs repository lets the manual-sync endpoint enqueue an email_sync job.
	service.SetJobsDispatcher(jobsRepository)
	handler := email.NewHandler(service)

	return &EmailContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

func newAccessControlContext(
	dbContext *database.DbContext,
	loggerService logger.LoggerServiceInterface,
) *AccessControlContext {
	repository := accesscontrol.NewRepository(dbContext)
	service := accesscontrol.NewService(repository)
	handler := accesscontrol.NewHandler(service, loggerService)

	return &AccessControlContext{
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

// newAIStack wires the persisted provider configuration to a hot-swappable
// AI service. The returned ai.ServiceInterface is an *ai.Manager: editing a
// provider through the API rebuilds the underlying router without a restart.
func newAIStack(dbContext *database.DbContext) (ai.ServiceInterface, *AIProvidersContext) {
	cfg := ai.LoadConfig()

	repository := aiproviders.NewRepository(dbContext)
	service := aiproviders.NewService(repository, cfg)

	manager := ai.NewManager(nil)

	if dbContext != nil && dbContext.GetDatabase() != nil {
		if err := service.EnsureDefaults(); err != nil {
			log.Printf("AI providers: failed to seed defaults: %v\n", err)
		}
	}

	rebuild := func() {
		models, err := service.GetProviderModels()
		if err != nil {
			log.Printf("AI providers: failed to load configuration: %v\n", err)
			return
		}
		manager.Swap(buildAIServiceFromModels(models, cfg))
	}

	if dbContext != nil && dbContext.GetDatabase() != nil {
		rebuild()
	}
	service.SetOnChange(rebuild)

	handler := aiproviders.NewHandler(service)

	return manager, &AIProvidersContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}

// newOllamaContext builds the Ollama management module. The daemon base URL is
// resolved dynamically from the persisted provider configuration so changes
// made through the UI take effect without a restart.
func newOllamaContext(aiProvidersService aiproviders.ServiceInterface, jobsRepository jobs.RepositoryInterface) *OllamaContext {
	cfg := ai.LoadConfig()
	fallbackBaseURL := cfg.OllamaBaseURL

	resolver := func() string {
		if aiProvidersService != nil {
			if models, err := aiProvidersService.GetProviderModels(); err == nil {
				for _, model := range models {
					if model.Name == aiproviders.ProviderOllama && strings.TrimSpace(model.BaseURL) != "" {
						return model.BaseURL
					}
				}
			}
		}
		return fallbackBaseURL
	}

	// enabled reads the live provider config so the autostart only fires when the
	// operator actually turned Ollama on (matches buildAIServiceFromModels).
	enabled := func() bool {
		if aiProvidersService == nil {
			return false
		}
		models, err := aiProvidersService.GetProviderModels()
		if err != nil {
			return false
		}
		for _, model := range models {
			if model.Name == aiproviders.ProviderOllama {
				return model.Enabled
			}
		}
		return false
	}

	service := ollamamgmt.NewService(resolver, jobsRepository)
	handler := ollamamgmt.NewHandler(service)
	daemon := ollamamgmt.NewDaemonManager(ollamamgmt.DaemonConfig{
		Autostart:    cfg.OllamaAutostart,
		Binary:       cfg.OllamaBinary,
		StartTimeout: cfg.OllamaStartTimeout,
	}, resolver, enabled)

	return &OllamaContext{
		Handler: handler,
		Service: service,
		Daemon:  daemon,
	}
}

func aiTaskTypes() []ai.TaskType {
	return []ai.TaskType{
		ai.TaskClassification,
		ai.TaskExtraction,
		ai.TaskSummarization,
		ai.TaskGeneration,
		ai.TaskSimple,
		ai.TaskComplex,
	}
}

func providerTimeout(model aiproviders.ProviderModel, cfg ai.Config) time.Duration {
	if model.Params.TimeoutSeconds > 0 {
		return time.Duration(model.Params.TimeoutSeconds) * time.Second
	}
	return cfg.DefaultTimeout
}

// withProviderRetry wraps a provider with its own persisted retry policy, so
// each provider's timeout/retry tuning (from the ai_providers table) is applied
// independently.
func withProviderRetry(provider ai.Provider, model aiproviders.ProviderModel) ai.Provider {
	backoff := time.Duration(model.Params.RetryBackoffMS) * time.Millisecond
	return ai.WithRetry(provider, model.Params.MaxRetries, backoff)
}

// buildAIServiceFromModels constructs the provider chain from persisted
// configuration. Operational tuning (model, base_url, timeout, retries) comes
// from the ai_providers table; only the API keys come from the environment.
// Providers are already ordered by priority; the first enabled one becomes
// primary and the rest are fallbacks. Cloud providers are skipped when their
// API key is missing. Returns nil when nothing is enabled.
func buildAIServiceFromModels(models []aiproviders.ProviderModel, cfg ai.Config) ai.ServiceInterface {
	var providers []ai.Provider

	for _, model := range models {
		if !model.Enabled {
			continue
		}

		switch model.Name {
		case aiproviders.ProviderOllama:
			keepAlive := model.Params.KeepAlive
			if keepAlive == "" {
				keepAlive = cfg.OllamaKeepAlive
			}
			base := ollama.NewProvider(model.BaseURL, model.Model, keepAlive, providerTimeout(model, cfg))
			providers = append(providers, withProviderRetry(base, model))
			log.Printf("AI provider enabled: ollama (%s @ %s)\n", model.Model, model.BaseURL)
		case aiproviders.ProviderOpenAI:
			if cfg.OpenAIAPIKey == "" {
				log.Println("AI provider openai enabled but no API key configured; skipping")
				continue
			}
			base := openai.NewProvider(cfg.OpenAIAPIKey, model.Model, model.BaseURL, providerTimeout(model, cfg))
			providers = append(providers, withProviderRetry(base, model))
			log.Printf("AI provider enabled: openai (%s)\n", model.Model)
		case aiproviders.ProviderAnthropic:
			if cfg.AnthropicAPIKey == "" {
				log.Println("AI provider anthropic enabled but no API key configured; skipping")
				continue
			}
			base := anthropic.NewProvider(cfg.AnthropicAPIKey, model.Model, providerTimeout(model, cfg))
			providers = append(providers, withProviderRetry(base, model))
			log.Printf("AI provider enabled: anthropic (%s)\n", model.Model)
		}
	}

	if len(providers) == 0 {
		log.Println("AI service has no enabled providers")
		return nil
	}

	router := ai.NewRouter()
	for _, taskType := range aiTaskTypes() {
		router.RegisterChain(taskType, providers...)
	}

	log.Printf("AI service enabled with %d provider(s)\n", len(providers))
	return ai.NewService(router)
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

// distributionDownloadsDir is where pre-built client apps (APKs, the browser
// extension zip) and their manifest.json are shipped, next to ./dist. CI builds
// the artifacts and the updater syncs this directory in place; the server only
// hosts them, it never builds them.
const distributionDownloadsDir = "./downloads"

// newDistributionContext wires the client-app download catalog. It is
// filesystem-backed (no DB): a missing downloads/ directory simply yields an
// empty catalog, so the dev server, which ships no artifacts, degrades cleanly.
func newDistributionContext() *DistributionContext {
	repository := distribution.NewRepository(distributionDownloadsDir)
	service := distribution.NewService(repository)
	handler := distribution.NewHandler(service)
	return &DistributionContext{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
