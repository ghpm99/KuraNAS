package engine

import (
	"nas-go/api/internal/worker/job"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	ollamaapi "nas-go/api/internal/api/v1/ollama"
	"nas-go/api/pkg/database"
)

type ollamaPullProgress struct {
	Status    string `json:"status"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Error     string `json:"error"`
}

// executeOllamaPullStep streams a model download from the Ollama daemon,
// reporting progress on the job step as the layers are fetched.
func executeOllamaPullStep(context *WorkerContext, step jobs.StepModel) error {
	var payload ollamaapi.PullStepPayload
	if err := json.Unmarshal(step.Payload, &payload); err != nil {
		return fmt.Errorf("decode ollama pull payload: %w", err)
	}
	if payload.Model == "" {
		return fmt.Errorf("ollama pull payload model is required")
	}
	baseURL := strings.TrimRight(payload.BaseURL, "/")
	if baseURL == "" {
		return fmt.Errorf("ollama pull payload base_url is required")
	}

	body, err := json.Marshal(map[string]any{"model": payload.Model, "stream": true})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(stdContext(), http.MethodPost, baseURL+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0} // model pulls can take several minutes
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama pull request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama pull returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lastReported := -1
	lastReportAt := time.Time{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var progress ollamaPullProgress
		if err := json.Unmarshal([]byte(line), &progress); err != nil {
			continue
		}
		if progress.Error != "" {
			return fmt.Errorf("ollama pull error: %s", progress.Error)
		}

		if progress.Total > 0 {
			pct := int(progress.Completed * 100 / progress.Total)
			if shouldReportProgress(pct, lastReported, lastReportAt) {
				reportPullProgress(context, step, pct)
				lastReported = pct
				lastReportAt = time.Now()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ollama pull stream read: %w", err)
	}

	return nil
}

// shouldReportProgress throttles DB writes: report on a >=1% change but no more
// than once per second, and always allow the first report.
func shouldReportProgress(pct, lastReported int, lastReportAt time.Time) bool {
	if lastReported < 0 {
		return true
	}
	if pct <= lastReported {
		return false
	}
	if pct-lastReported >= 1 && time.Since(lastReportAt) >= time.Second {
		return true
	}
	return pct >= 100
}

func reportPullProgress(context *WorkerContext, step jobs.StepModel, pct int) {
	if context == nil || context.JobsRepository == nil {
		return
	}
	_ = database.ExecOptionalTx(context.JobsRepository.GetDbContext(), func(tx *sql.Tx) error {
		_, err := context.JobsRepository.UpdateStepExecution(tx, step.ID, string(job.StepStatusRunning), pct, step.Attempts+1, nil, nil, nil)
		return err
	})
}

func stdContext() context.Context {
	return context.Background()
}
