package email

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/mailfetch"
	"nas-go/api/pkg/utils"
)

// Job/step type strings, duplicated as literals (like the ollama domain) to
// avoid importing the worker package and creating an import cycle. They must
// match the StepType/JobType values enumerated in internal/worker/job.
const (
	syncJobType        = "email_sync"
	fetchStepType      = "email_fetch"
	prefilterStepType  = "email_prefilter"
	analyzeStepType    = "email_analyze"
	perAccountFetchTTL = 60 * time.Second
)

// SyncStats summarizes one fetch pass for the worker's notification logic.
type SyncStats struct {
	Accounts       int
	Fetched        int
	Failures       int
	ReauthRequired []string
}

// SyncEnabledAccounts fetches and stores new messages for every sync-enabled
// account. A token rejected during refresh marks that account reauth_required
// (handled inside ValidAccessToken) and is reported back without aborting the
// other accounts.
func (s *Service) SyncEnabledAccounts() (SyncStats, error) {
	accounts, err := s.repository.ListAccounts()
	if err != nil {
		return SyncStats{}, err
	}

	var stats SyncStats
	for _, account := range accounts {
		if !account.SyncEnabled {
			continue
		}
		stats.Accounts++

		fetched, syncErr := s.syncAccount(account)
		switch {
		case errors.Is(syncErr, ErrReauthRequired):
			stats.ReauthRequired = append(stats.ReauthRequired, account.Address)
		case syncErr != nil:
			stats.Failures++
			log.Printf("[email] sync failed for account %d (%s): %v\n", account.ID, account.Address, syncErr)
		default:
			stats.Fetched += fetched
		}
	}
	return stats, nil
}

func (s *Service) syncAccount(account AccountModel) (int, error) {
	fetcher, ok := s.fetchers[account.Provider]
	if !ok || fetcher == nil {
		return 0, fmt.Errorf("email: no fetcher for provider %q", account.Provider)
	}

	token, err := s.ValidAccessToken(account.ID)
	if err != nil {
		return 0, err
	}

	since := time.Time{}
	if account.LastSyncAt != nil {
		since = *account.LastSyncAt
	}

	// Capture the cursor before fetching: messages that arrive mid-fetch are
	// caught next pass, and the UNIQUE(account, provider_message_id) constraint
	// makes any boundary overlap idempotent.
	syncStart := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), perAccountFetchTTL)
	defer cancel()

	rawMessages, err := fetcher.ListNewMessages(ctx, token, since, s.maxPerAccount)
	if err != nil {
		return 0, err
	}

	inserted := 0
	for _, raw := range rawMessages {
		stored, insertErr := s.repository.InsertMessage(s.toMessageModel(account.ID, raw))
		if insertErr != nil {
			return inserted, insertErr
		}
		if stored {
			inserted++
		}
	}

	if err := s.repository.UpdateAccountLastSync(account.ID, syncStart); err != nil {
		return inserted, err
	}
	return inserted, nil
}

// toMessageModel converts a raw provider message into the stored shape: the
// body is sanitized to plain text, URLs become bare link domains (never
// visited) and attachments keep only metadata.
func (s *Service) toMessageModel(accountID int, raw mailfetch.RawMessage) MessageModel {
	body, snippet := sanitizeBody(raw.Body, raw.BodyIsHTML)

	receivedAt := raw.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now()
	}

	attachments := make([]AttachmentMeta, 0, len(raw.Attachments))
	for _, attachment := range raw.Attachments {
		attachments = append(attachments, AttachmentMeta{
			Filename: attachment.Filename,
			Mime:     attachment.Mime,
			Size:     attachment.Size,
		})
	}

	return MessageModel{
		AccountID:         accountID,
		ProviderMessageID: raw.ProviderMessageID,
		SenderName:        raw.SenderName,
		SenderAddress:     raw.SenderAddress,
		Subject:           raw.Subject,
		Snippet:           snippet,
		SanitizedBody:     body,
		ReceivedAt:        receivedAt,
		AuthResults:       AuthResults(raw.AuthResults),
		Attachments:       attachments,
		LinkDomains:       extractLinkDomains(raw.Body),
		Status:            MsgStatusPending,
	}
}

// PrefilterPending runs the deterministic pre-filter over pending messages,
// flagging obvious spam/phishing so it never reaches the LLM (task 16).
func (s *Service) PrefilterPending() (int, error) {
	pending, err := s.repository.ListPendingMessages(pendingScanLimit)
	if err != nil {
		return 0, err
	}

	flagged := 0
	for _, message := range pending {
		status, rules := prefilter(message)
		if status == MsgStatusPending {
			continue
		}
		if err := s.repository.UpdateMessagePrefilter(message.ID, status, rules); err != nil {
			return flagged, err
		}
		flagged++
	}
	return flagged, nil
}

// PurgeExpired drops messages older than the retention window.
func (s *Service) PurgeExpired() (int, error) {
	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	return s.repository.PurgeMessagesBefore(cutoff)
}

// ListMessages returns one lean page of synced messages (no body), clamping the
// page size so a client cannot ask for an unbounded payload.
func (s *Service) ListMessages(page, pageSize int) (utils.PaginationResponse[MessageDto], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	models, err := s.repository.ListMessages(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[MessageDto]{}, err
	}

	dtos := make([]MessageDto, 0, len(models.Items))
	for _, model := range models.Items {
		dtos = append(dtos, model.toDto())
	}
	return utils.PaginationResponse[MessageDto]{Items: dtos, Pagination: models.Pagination}, nil
}

// EnqueueSync queues one email_sync job (manual trigger). It validates the
// account first so a stale id 404s; the job itself re-syncs every enabled
// account, which is idempotent and cheap.
func (s *Service) EnqueueSync(accountID int) (int, error) {
	if _, err := s.repository.GetAccountByID(accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, err
	}
	if s.jobsRepo == nil {
		return 0, ErrSyncUnavailable
	}

	var jobID int
	err := s.jobsRepo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		createdJob, createErr := s.jobsRepo.CreateJob(tx, jobs.JobModel{
			Type:     syncJobType,
			Priority: "normal",
			Status:   "queued",
		})
		if createErr != nil {
			return createErr
		}

		fetchStep, stepErr := s.jobsRepo.CreateStep(tx, jobs.StepModel{
			JobID:       createdJob.ID,
			Type:        fetchStepType,
			Status:      "queued",
			DependsOn:   []byte("[]"),
			MaxAttempts: 1,
		})
		if stepErr != nil {
			return stepErr
		}

		dependsOn, marshalErr := json.Marshal([]int{fetchStep.ID})
		if marshalErr != nil {
			return marshalErr
		}
		prefilterStep, stepErr := s.jobsRepo.CreateStep(tx, jobs.StepModel{
			JobID:       createdJob.ID,
			Type:        prefilterStepType,
			Status:      "queued",
			DependsOn:   dependsOn,
			MaxAttempts: 1,
		})
		if stepErr != nil {
			return stepErr
		}

		analyzeDependsOn, marshalErr := json.Marshal([]int{prefilterStep.ID})
		if marshalErr != nil {
			return marshalErr
		}
		if _, stepErr := s.jobsRepo.CreateStep(tx, jobs.StepModel{
			JobID:       createdJob.ID,
			Type:        analyzeStepType,
			Status:      "queued",
			DependsOn:   analyzeDependsOn,
			MaxAttempts: 1,
		}); stepErr != nil {
			return stepErr
		}

		jobID = createdJob.ID
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("EnqueueSync: %w", err)
	}
	return jobID, nil
}
