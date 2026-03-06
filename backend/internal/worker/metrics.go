package worker

import (
	"sync"
	"time"

	"nas-go/api/internal/worker/domain"
)

type SchedulerMetricsSnapshot struct {
	TotalJobsSeen      int64                       `json:"total_jobs_seen"`
	QueuedJobs         int                         `json:"queued_jobs"`
	RunningJobs        int                         `json:"running_jobs"`
	StepExecByStatus   map[domain.StepStatus]int64 `json:"step_exec_by_status"`
	StepDurationMillis map[domain.StepType]float64 `json:"avg_step_duration_millis"`
	LastUpdatedAt      time.Time                   `json:"last_updated_at"`
}

type schedulerStepDurationAggregate struct {
	totalMillis int64
	count       int64
}

type SchedulerMetrics struct {
	mu                    sync.RWMutex
	totalJobsSeen         int64
	queuedJobs            int
	runningJobs           int
	stepExecByStatus      map[domain.StepStatus]int64
	stepDurationByTypeRaw map[domain.StepType]schedulerStepDurationAggregate
	lastUpdatedAt         time.Time
}

func NewSchedulerMetrics() *SchedulerMetrics {
	return &SchedulerMetrics{
		stepExecByStatus:      map[domain.StepStatus]int64{},
		stepDurationByTypeRaw: map[domain.StepType]schedulerStepDurationAggregate{},
		lastUpdatedAt:         time.Now().UTC(),
	}
}

func (m *SchedulerMetrics) RecordJobQueue(queued int, running int) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.queuedJobs = queued
	m.runningJobs = running
	m.totalJobsSeen += int64(queued + running)
	m.lastUpdatedAt = time.Now().UTC()
}

func (m *SchedulerMetrics) RecordStepExecution(stepType domain.StepType, status domain.StepStatus, duration time.Duration) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.stepExecByStatus[status]++
	durationAggregate := m.stepDurationByTypeRaw[stepType]
	durationAggregate.totalMillis += duration.Milliseconds()
	durationAggregate.count++
	m.stepDurationByTypeRaw[stepType] = durationAggregate
	m.lastUpdatedAt = time.Now().UTC()
}

func (m *SchedulerMetrics) Snapshot() SchedulerMetricsSnapshot {
	if m == nil {
		return SchedulerMetricsSnapshot{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	stepExecByStatus := make(map[domain.StepStatus]int64, len(m.stepExecByStatus))
	for status, total := range m.stepExecByStatus {
		stepExecByStatus[status] = total
	}

	stepDurationMillis := make(map[domain.StepType]float64, len(m.stepDurationByTypeRaw))
	for stepType, aggregate := range m.stepDurationByTypeRaw {
		if aggregate.count == 0 {
			stepDurationMillis[stepType] = 0
			continue
		}
		stepDurationMillis[stepType] = float64(aggregate.totalMillis) / float64(aggregate.count)
	}

	return SchedulerMetricsSnapshot{
		TotalJobsSeen:      m.totalJobsSeen,
		QueuedJobs:         m.queuedJobs,
		RunningJobs:        m.runningJobs,
		StepExecByStatus:   stepExecByStatus,
		StepDurationMillis: stepDurationMillis,
		LastUpdatedAt:      m.lastUpdatedAt,
	}
}
