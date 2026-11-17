package engine

import (
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/applog"
	"time"
)

const (
	takeoutProgressEntryInterval = 50
	takeoutProgressTimeInterval  = time.Second
)

type takeoutProgressReporter struct {
	context             *WorkerContext
	step                jobs.StepModel
	now                 func() time.Time
	lastReportedAt      time.Time
	lastReportedEntries int
	hasReported         bool
}

func newTakeoutProgressReporter(context *WorkerContext, step jobs.StepModel) *takeoutProgressReporter {
	return &takeoutProgressReporter{context: context, step: step, now: time.Now}
}

func (r *takeoutProgressReporter) report(processedEntries int, totalEntries int) {
	if !r.shouldReport(processedEntries, totalEntries) {
		return
	}
	r.hasReported = true
	r.lastReportedEntries = processedEntries
	r.lastReportedAt = r.now()

	reportStepProgress(r.context, r.step, takeoutProgressPercent(processedEntries, totalEntries))
	r.recordImportProgress(processedEntries, totalEntries)
}

func (r *takeoutProgressReporter) shouldReport(processedEntries int, totalEntries int) bool {
	if !r.hasReported || processedEntries >= totalEntries {
		return true
	}
	if processedEntries-r.lastReportedEntries >= takeoutProgressEntryInterval {
		return true
	}
	return r.now().Sub(r.lastReportedAt) >= takeoutProgressTimeInterval
}

func (r *takeoutProgressReporter) recordImportProgress(processedEntries int, totalEntries int) {
	if r.context.TakeoutImports == nil {
		return
	}
	if err := r.context.TakeoutImports.UpdateImportProgress(r.step.JobID, processedEntries, totalEntries); err != nil {
		applog.Error("takeout import record: update progress failed", "job_id", r.step.JobID, "error", err.Error())
	}
}

func takeoutProgressPercent(processedEntries int, totalEntries int) int {
	if totalEntries <= 0 {
		return 100
	}
	return processedEntries * 100 / totalEntries
}
