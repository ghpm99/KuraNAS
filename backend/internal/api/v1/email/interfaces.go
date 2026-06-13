package email

import (
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	ListAccounts() ([]AccountModel, error)
	GetAccountByID(id int) (AccountModel, error)
	UpsertAccount(model AccountModel) (int, error)
	UpdateAccountTokens(id int, tokenCiphertext []byte, status AccountStatus, lastError string) error
	UpdateSyncEnabled(id int, enabled bool) error
	UpdateAccountLastSync(id int, syncedAt time.Time) error
	DeleteAccount(id int) error
	InsertMessage(message MessageModel) (inserted bool, err error)
	ListMessages(page, pageSize int) (utils.PaginationResponse[MessageModel], error)
	ListPendingMessages(limit int) ([]MessageModel, error)
	UpdateMessagePrefilter(id int, status MessageStatus, rules []string) error
	PurgeMessagesBefore(cutoff time.Time) (int, error)
	ListMessagesForAnalysis(limit int) ([]MessageModel, error)
	UpsertAnalysis(model AnalysisModel) error
	UpdateMessageAnalyzed(id int, status MessageStatus) error
	GetAnalysisByMessage(messageID int) (AnalysisModel, error)
	GetProviderPreference() (string, error)
	SetProviderPreference(value string) error
}

type ServiceInterface interface {
	ListAccounts() ([]AccountDto, error)
	DeleteAccount(id int) error
	SetSyncEnabled(id int, enabled bool) error
	GoogleAuthURL() (GoogleAuthURLDto, error)
	HandleGoogleCallback(state string, code string) error
	StartMicrosoftDeviceCode() (DeviceCodeDto, error)
	MicrosoftDeviceCodeStatus() DeviceCodeStatusDto
	// ValidAccessToken returns a usable access token for the account,
	// refreshing (and re-sealing) it when expired.
	ValidAccessToken(accountID int) (string, error)
	// ListMessages returns one lean page of synced messages (no body).
	ListMessages(page, pageSize int) (utils.PaginationResponse[MessageDto], error)
	// EnqueueSync queues an email_sync job (manual trigger), after checking the
	// account exists. Returns the job id.
	EnqueueSync(accountID int) (int, error)
	// GetMessageAnalysis returns one message's stored AI verdict/summary.
	GetMessageAnalysis(messageID int) (AnalysisDto, error)
	// GetProviderPreference / SetProviderPreference read and set which AI
	// provider analyzes e-mail.
	GetProviderPreference() (ProviderPreferenceDto, error)
	SetProviderPreference(provider string) (ProviderPreferenceDto, error)
	WorkerInterface
}

// WorkerInterface is the narrow capability the sync worker consumes: fetch new
// messages for every enabled account, pre-filter what is pending, and purge by
// retention. It is split out so the worker depends on a tiny seam.
type WorkerInterface interface {
	// SyncEnabledAccounts fetches+stores new messages for every sync-enabled
	// account, returning per-run stats (including accounts that need reauth).
	SyncEnabledAccounts() (SyncStats, error)
	// PrefilterPending runs the deterministic pre-filter over pending messages,
	// flagging spam. Returns how many were flagged.
	PrefilterPending() (int, error)
	// PurgeExpired drops messages past the retention window. Returns the count.
	PurgeExpired() (int, error)
	// AnalyzePending classifies + summarizes pending messages, returning the
	// detections worth notifying. AIUnavailable in the stats means messages were
	// left pending for the next cycle.
	AnalyzePending() (AnalyzeStats, error)
}
