package email

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"nas-go/api/pkg/applog"
)

// AIRouter is the narrow AI capability the analysis consumes: run the default
// routing chain (Execute) or pin one provider by name (Named). It is the seam
// over *ai.Manager so the e-mail domain depends on a tiny interface, never the
// whole manager. The pipeline calls ONLY Provider.Complete / Service.Execute —
// never pkg/ai/agent — so an adversarial e-mail can never trigger a tool.
type AIRouter interface {
	Execute(ctx context.Context, req ai.Request) (ai.Response, error)
	Named(name string) ai.Provider
	Enabled() bool
}

// Provider-preference values stored under the email_ai_provider setting. "auto"
// uses the default router chain; the others pin a single named provider. Default
// is local Ollama: sending private mail to a cloud model is an explicit choice.
const (
	ProviderPrefAuto      = "auto"
	ProviderPrefOllama    = "ollama"
	ProviderPrefOpenAI    = "openai"
	ProviderPrefAnthropic = "anthropic"
	defaultProviderPref   = ProviderPrefOllama
)

const (
	// analysisScanLimit caps how many pending messages one analysis pass handles.
	analysisScanLimit = 200
	// maxBodyForAnalysis bounds the body handed to the LLM (hard rule 3).
	maxBodyForAnalysis = 16 * 1024
	// maxEvidenceItems caps the evidence list persisted/shown.
	maxEvidenceItems = 6
	// defaultSuspiciousRisk is the risk score for a fail-closed verdict.
	defaultSuspiciousRisk = 50
	// evidenceParseFailed marks a verdict produced by the fail-closed path.
	evidenceParseFailed = "ANALYSIS_PARSE_FAILED"
)

var errProviderUnavailable = errors.New("email: selected AI provider is not enabled")

// EmailDetection identifies a message that warrants a notification.
type EmailDetection struct {
	AccountID int
	Subject   string
}

// AnalyzeStats summarizes one analysis pass for the worker's notification and
// step-status logic. AIUnavailable means no provider answered (or the chosen one
// is off): the step reports partial_fail and the messages stay pending for the
// next cycle.
type AnalyzeStats struct {
	Analyzed      int
	AIUnavailable bool
	Malicious     []EmailDetection
	Suspicious    []EmailDetection
	Important     []EmailDetection
}

// SetAIRouter wires the AI seam used by the analysis. The composition root calls
// it with an adapter over the hot-swappable ai.Manager; without it AnalyzePending
// reports the AI as unavailable and leaves every message pending.
func (s *Service) SetAIRouter(router AIRouter) {
	s.aiRouter = router
}

// AnalyzePending classifies every pending message (verdict + risk + evidence +
// importance), summarizes the legitimate ones, persists the verdict, expunges the
// sanitized body, and returns the detections for the worker to notify. It is the
// most sensitive point of the threat model: the e-mail body is adversarial input,
// so the model output is treated strictly as data (validated, never re-fed) and
// any parse failure is fail-closed to "suspicious".
func (s *Service) AnalyzePending() (AnalyzeStats, error) {
	if s.aiRouter == nil || !s.aiRouter.Enabled() {
		return AnalyzeStats{AIUnavailable: true}, nil
	}

	pref := s.resolveProviderPreference()
	// A pinned provider that is not currently enabled is treated as unavailable:
	// we never silently fall back to a different provider (privacy decision).
	if pref != ProviderPrefAuto && s.aiRouter.Named(pref) == nil {
		return AnalyzeStats{AIUnavailable: true}, nil
	}

	pending, err := s.repository.ListMessagesForAnalysis(analysisScanLimit)
	if err != nil {
		return AnalyzeStats{}, err
	}

	var stats AnalyzeStats
	for _, message := range pending {
		analysis, analyzeErr := s.analyzeMessage(message, pref)
		if analyzeErr != nil {
			// The provider failed or vanished mid-pass: stop here, leave this and
			// the remaining messages pending, and let the next cycle retry. No
			// crash, no infinite retry of a single message in one pass.
			applog.Warn("email analysis aborted: provider unavailable",
				"id", message.ID, "error", analyzeErr.Error())
			stats.AIUnavailable = true
			return stats, nil
		}

		if err := s.repository.UpsertAnalysis(analysis); err != nil {
			return stats, err
		}
		// Retention rule A7: drop the body now that a verdict exists.
		if err := s.repository.UpdateMessageAnalyzed(message.ID, MsgStatusAnalyzed); err != nil {
			return stats, err
		}
		stats.Analyzed++
		recordDetection(&stats, message, analysis)
	}
	return stats, nil
}

// analyzeMessage runs the classification (and, for legitimate mail, the summary)
// for one message. The returned error is reserved for AI unavailability — a model
// answer that does not match the schema is NOT an error, it is fail-closed to a
// suspicious verdict inside parseClassification.
func (s *Service) analyzeMessage(message MessageModel, pref string) (AnalysisModel, error) {
	nonce, err := randomToken(8)
	if err != nil {
		return AnalysisModel{}, err
	}

	body := truncateBody(message.SanitizedBody)
	userPrompt := prompts.EmailClassificationUserPrompt(nonce, buildEvidenceBlock(message), message.Subject, body)

	resp, err := s.complete(context.Background(), pref, ai.Request{
		TaskType:     ai.TaskClassification,
		SystemPrompt: prompts.EmailClassificationSystemPrompt(),
		Prompt:       userPrompt,
		MaxTokens:    500,
		Temperature:  0,
	})
	if err != nil {
		return AnalysisModel{}, err
	}

	analysis := parseClassification(resp.Content)
	analysis.MessageID = message.ID
	analysis.ProviderUsed = resp.Provider
	analysis.ModelUsed = resp.Model

	if analysis.Verdict == VerdictLegitimate {
		// The summary is best-effort: its failure must not undo a good
		// classification, so a summary error only logs and leaves it empty.
		analysis.Summary = s.summarize(message, pref)
	}
	return analysis, nil
}

// summarize asks the model for a 2-3 sentence summary of a legitimate message,
// with a fresh per-request delimiter nonce. Any failure (transport or parse)
// yields an empty summary; the classification has already succeeded.
func (s *Service) summarize(message MessageModel, pref string) string {
	nonce, err := randomToken(8)
	if err != nil {
		return ""
	}

	prompt := prompts.EmailSummaryUserPrompt(nonce, message.Subject, truncateBody(message.SanitizedBody))
	resp, err := s.complete(context.Background(), pref, ai.Request{
		TaskType:     ai.TaskSummarization,
		SystemPrompt: prompts.EmailSummarySystemPrompt(),
		Prompt:       prompt,
		MaxTokens:    300,
		Temperature:  0.2,
	})
	if err != nil {
		applog.Warn("email summary failed", "id", message.ID, "error", err.Error())
		return ""
	}
	return parseSummary(resp.Content)
}

// complete runs one completion, pinned to the chosen provider or through the
// default chain. A pinned-but-disabled provider is an unavailability error so the
// caller leaves the message pending.
func (s *Service) complete(ctx context.Context, pref string, req ai.Request) (ai.Response, error) {
	if pref == "" || pref == ProviderPrefAuto {
		return s.aiRouter.Execute(ctx, req)
	}
	provider := s.aiRouter.Named(pref)
	if provider == nil {
		return ai.Response{}, errProviderUnavailable
	}
	return provider.Complete(ctx, req)
}

// resolveProviderPreference reads the stored provider choice, falling back to the
// local default on a read error or an unrecognized value.
func (s *Service) resolveProviderPreference() string {
	pref, err := s.repository.GetProviderPreference()
	if err != nil {
		applog.Warn("email: could not read AI provider preference, using default", "error", err.Error())
		return defaultProviderPref
	}
	pref = strings.ToLower(strings.TrimSpace(pref))
	if isValidProviderPref(pref) {
		return pref
	}
	return defaultProviderPref
}

// GetProviderPreference returns the configured AI provider for e-mail analysis
// (defaulting to local Ollama when unset), for the settings endpoint.
func (s *Service) GetProviderPreference() (ProviderPreferenceDto, error) {
	pref, err := s.repository.GetProviderPreference()
	if err != nil {
		return ProviderPreferenceDto{}, err
	}
	pref = strings.ToLower(strings.TrimSpace(pref))
	if pref == "" {
		pref = defaultProviderPref
	}
	return ProviderPreferenceDto{Provider: pref}, nil
}

// SetProviderPreference validates and stores the AI provider for e-mail analysis.
// The change takes effect on the next pass with no restart (the analysis reads
// the live preference and the manager hot-swaps providers).
func (s *Service) SetProviderPreference(provider string) (ProviderPreferenceDto, error) {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if !isValidProviderPref(normalized) {
		return ProviderPreferenceDto{}, ErrInvalidProvider
	}
	if err := s.repository.SetProviderPreference(normalized); err != nil {
		return ProviderPreferenceDto{}, err
	}
	return ProviderPreferenceDto{Provider: normalized}, nil
}

// GetMessageAnalysis returns one message's stored verdict for the summary
// endpoint, or ErrAnalysisNotFound when it has not been analyzed.
func (s *Service) GetMessageAnalysis(messageID int) (AnalysisDto, error) {
	model, err := s.repository.GetAnalysisByMessage(messageID)
	if errors.Is(err, sql.ErrNoRows) {
		return AnalysisDto{}, ErrAnalysisNotFound
	}
	if err != nil {
		return AnalysisDto{}, err
	}
	return model.toDto(), nil
}

func recordDetection(stats *AnalyzeStats, message MessageModel, analysis AnalysisModel) {
	detection := EmailDetection{AccountID: message.AccountID, Subject: message.Subject}
	switch analysis.Verdict {
	case VerdictMalicious:
		stats.Malicious = append(stats.Malicious, detection)
	case VerdictSuspicious:
		stats.Suspicious = append(stats.Suspicious, detection)
	case VerdictLegitimate:
		if analysis.Importance == ImportanceHigh {
			stats.Important = append(stats.Important, detection)
		}
	}
}

// buildEvidenceBlock renders the trusted deterministic signals fed to the model
// alongside (but clearly separated from) the untrusted e-mail body.
func buildEvidenceBlock(m MessageModel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "- Sender address: %s\n", orNA(m.SenderAddress))
	fmt.Fprintf(&b, "- Authentication: SPF=%s, DKIM=%s, DMARC=%s\n",
		orNA(m.AuthResults.SPF), orNA(m.AuthResults.DKIM), orNA(m.AuthResults.DMARC))

	senderDomain := domainOf(m.SenderAddress)
	if senderDomain != "" && len(m.LinkDomains) > 0 && !linkContainsDomain(m.LinkDomains, senderDomain) {
		fmt.Fprintf(&b, "- Links point to domains other than the sender's (%s): %s\n",
			senderDomain, strings.Join(m.LinkDomains, ", "))
	} else if len(m.LinkDomains) > 0 {
		fmt.Fprintf(&b, "- Link domains: %s\n", strings.Join(m.LinkDomains, ", "))
	}

	if len(m.Attachments) > 0 {
		names := make([]string, 0, len(m.Attachments))
		for _, attachment := range m.Attachments {
			names = append(names, fmt.Sprintf("%s (%s)", attachment.Filename, attachment.Mime))
		}
		fmt.Fprintf(&b, "- Attachments (metadata only, never opened): %s\n", strings.Join(names, "; "))
	}

	if len(m.PrefilterRules) > 0 {
		fmt.Fprintf(&b, "- Deterministic pre-filter flags: %s\n", strings.Join(m.PrefilterRules, ", "))
	}

	return strings.TrimRight(b.String(), "\n")
}

func orNA(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func isValidProviderPref(p string) bool {
	switch p {
	case ProviderPrefAuto, ProviderPrefOllama, ProviderPrefOpenAI, ProviderPrefAnthropic:
		return true
	default:
		return false
	}
}

func truncateBody(body string) string {
	if len(body) <= maxBodyForAnalysis {
		return body
	}
	return strings.ToValidUTF8(body[:maxBodyForAnalysis], "")
}
