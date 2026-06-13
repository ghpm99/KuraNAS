package email

import (
	"encoding/json"
	"strings"
)

// classificationResponse is the closed schema the classifier must emit. Unknown
// fields are rejected (DisallowUnknownFields), so an adversarial e-mail cannot
// smuggle extra keys past the validator.
type classificationResponse struct {
	Verdict    string   `json:"verdict"`
	RiskScore  int      `json:"risk_score"`
	Evidence   []string `json:"evidence"`
	Importance string   `json:"importance"`
}

type summaryResponse struct {
	Summary string `json:"summary"`
}

// parseClassification strictly decodes the model output into a verdict. ANY
// deviation from the closed schema — non-JSON, unknown field, invalid enum — is
// fail-closed to a "suspicious" verdict with the ANALYSIS_PARSE_FAILED marker.
// The e-mail body is adversarial input, so the only thing it can ever achieve is
// a wrong-but-safe classification; it can never push the output off contract.
func parseClassification(content string) AnalysisModel {
	failClosed := AnalysisModel{
		Verdict:    VerdictSuspicious,
		RiskScore:  defaultSuspiciousRisk,
		Evidence:   []string{evidenceParseFailed},
		Importance: ImportanceNormal,
	}

	raw := extractJSONObject(content)
	if raw == "" {
		return failClosed
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var parsed classificationResponse
	if err := dec.Decode(&parsed); err != nil {
		return failClosed
	}

	verdict := Verdict(strings.ToLower(strings.TrimSpace(parsed.Verdict)))
	importance := Importance(strings.ToLower(strings.TrimSpace(parsed.Importance)))
	if !verdict.IsValid() || !importance.IsValid() {
		return failClosed
	}

	return AnalysisModel{
		Verdict:    verdict,
		RiskScore:  clampRisk(parsed.RiskScore),
		Evidence:   sanitizeEvidence(parsed.Evidence),
		Importance: importance,
	}
}

// parseSummary decodes the summary answer. A failure yields an empty string —
// the summary is best-effort and never blocks a completed classification.
func parseSummary(content string) string {
	raw := extractJSONObject(content)
	if raw == "" {
		return ""
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var parsed summaryResponse
	if err := dec.Decode(&parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Summary)
}

// extractJSONObject returns the substring from the first '{' to the last '}'
// (after stripping any Markdown fences), so surrounding prose a small model adds
// does not break parsing. It does NOT weaken the schema: the extracted text is
// still decoded with unknown-field rejection.
func extractJSONObject(content string) string {
	s := stripFences(content)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < 0 || end < start {
		return ""
	}
	return s[start : end+1]
}

// stripFences removes Markdown ``` fences a model may wrap its JSON in.
func stripFences(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	lines := strings.Split(trimmed, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func sanitizeEvidence(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
		if len(out) == maxEvidenceItems {
			break
		}
	}
	return out
}

func clampRisk(score int) int {
	switch {
	case score < 0:
		return 0
	case score > 100:
		return 100
	default:
		return score
	}
}
