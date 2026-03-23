## 1. Feature Overview
- Name: Jobs and Worker Processing Pipeline
- Summary: Asynchronous orchestration layer for scans/uploads/indexing tasks with step dependencies, retries, cancellation, scheduling, and filesystem watcher triggers.
- Target user: Indirect end-user capability (supports all async-heavy features).
- Business value: Keeps UI responsive while ensuring durable background processing.

## 2. Problem Statement
- Metadata extraction, checksums, thumbnails, and indexing are expensive.
- Without worker orchestration, synchronous requests would be slow and unreliable.

## 3. Current Behavior
- Workers start only when `ENABLE_WORKERS=true`.
- Job scheduler pulls queued jobs by priority and executes ready steps based on dependency graph.
- Step executors include scan filesystem, diff against DB, mark deleted, persist, metadata, checksum, thumbnail, playlist index.
- Watcher snapshots entry point and enqueues targeted jobs or full scan jobs based on change volume.
- Startup scan job is auto-enqueued when workers run.

## 4. User Flows
### 4.1 Main Flow
1. User triggers upload or scan-related action.
2. System creates worker job + steps in DB.
3. Scheduler executes steps respecting dependencies and attempts.
4. Job transitions to completed/partial_fail/failed/canceled.
5. User can inspect job status and steps via jobs API.

### 4.2 Alternative Flows
1. Filesystem watcher detects changes and enqueues targeted file jobs.
2. Legacy fallback pipeline (non-job channel tasks) runs when orchestrator is unavailable.

### 4.3 Error Scenarios
1. Step executor missing for step type -> step failure.
2. Dependency dead-end/no ready steps -> queued steps canceled.
3. Cancellation requested mid-job -> job and remaining steps move to canceled path.

## 5. Functional Requirements
- The system must persist jobs and steps with statuses and progress.
- The system must execute dependent steps in valid order.
- The user can list jobs, view steps, and request cancellation.
- The system should schedule jobs by priority and bounded concurrency.

## 6. Business Rules
- Job priorities: low, normal, high.
- Job statuses: queued, running, partial_fail, failed, completed, canceled.
- Step statuses: queued, running, completed, failed, canceled, skipped.
- Upload-derived job steps always start from persist step before dependent processing.
- High filesystem change volume triggers full scan fallback instead of many micro-jobs.

## 7. Data Model (Business View)
- Worker Job (`worker_job`): type, priority, scope, lifecycle timestamps, cancel flag, last_error.
- Worker Step (`worker_step`): job relation, step type, dependency list, attempts, progress, payload, status.
- Jobs API DTO adds computed progress summary.

## 8. Interfaces
- User interfaces: indirect (files upload/scan actions, notifications, any domain using async processing).
- APIs:
  - `GET /jobs`
  - `GET /jobs/:id`
  - `GET /jobs/:id/steps`
  - `POST /jobs/:id/cancel`

## 9. Dependencies
- Files service, metadata repository, video service.
- Notifications for lifecycle events.
- Environment settings for concurrency/backoff/poll intervals.

## 10. Limitations / Gaps
- Limited user-facing retry controls per failed step.
- Job visibility is API-level; no dedicated frontend jobs dashboard currently.

## 11. Opportunities
- Add UI monitor for job queue and step retries.
- Add typed remediation actions for failed steps.

## 12. Acceptance Criteria
- Given a planned job with dependencies, when created, then steps are persisted with dependency IDs.
- Given queued jobs exist, when scheduler polls, then highest priority queued jobs are enqueued first.
- Given cancel is requested for a running/queued job, when processed, then job eventually reaches canceled terminal state.
- Given upload processing job, when executed, then persist step runs before metadata/checksum/thumbnail steps.

## 13. Assumptions
- Worker DB and filesystem access are available and stable.
