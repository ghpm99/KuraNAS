## 1. Feature Overview

* Name: Jobs and Worker Orchestration
* Summary: Manages asynchronous background processing for filesystem scans, metadata extraction, checksums, thumbnails, and playlist indexing.
* Purpose: Decouple heavy processing from request lifecycle while keeping indexed data in sync with disk changes.
* Business value: Enables scalable ingestion and continuous synchronization for large media libraries.

## 2. Current Implementation

* How it works today: Worker system supports queued jobs and step orchestration with dependencies; API exposes job status/steps/cancel operations.
* Main flows:
  * Startup enqueues initial scan job.
  * Watcher snapshots entrypoint every 5s and enqueues targeted/full jobs on detected changes.
  * Orchestrator executes planned steps (`scan_filesystem`, `diff_against_db`, `mark_deleted`, plus file-specific steps).
  * Jobs endpoint provides progress and supports cancellation request.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/jobs`
  * `GET /api/v1/jobs/:id`
  * `GET /api/v1/jobs/:id/steps`
  * `POST /api/v1/jobs/:id/cancel`
  * Background entry: `worker.StartWorkers`, filesystem watcher loop, scheduler/orchestrator loops.
* Key files involved (list with paths):
  * `backend/internal/api/v1/jobs/handler.go`
  * `backend/internal/api/v1/jobs/service.go`
  * `backend/internal/api/v1/jobs/repository.go`
  * `backend/internal/worker/worker.go`
  * `backend/internal/worker/job_scheduler.go`
  * `backend/internal/worker/step_executors.go`
  * `backend/internal/worker/watcher.go`
  * `backend/internal/worker/fileProcessingPipeline.go`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Backend-only core orchestration plus frontend consumption through analytics/status pages.
* Data flow (step-by-step):
  * Job plan is created with type/priority/scope/steps.
  * Repository persists `worker_job` and `worker_step` rows.
  * Scheduler fetches runnable steps based on dependencies/status.
  * Step executor performs IO/DB work and updates progress/status/errors.
  * API reads persisted job/step state and computes progress for clients.
* External integrations:
  * Filesystem traversal, metadata scripts, ffmpeg/thumbnail flows (via dependent services).
* State management (if applicable):
  * Durable DB-backed queue/state model.

## 4. Data Model

* Entities involved:
  * Worker job and worker step with dependency graph semantics.
* Database tables / schemas:
  * `worker_job`
  * `worker_step`
* Relationships:
  * `worker_step.job_id -> worker_job.id`.
* Important fields:
  * `worker_job.type`, `priority`, `status`, `cancel_requested`.
  * `worker_step.type`, `depends_on`, `attempts`, `max_attempts`, `progress`, `payload`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Workers run only when `ENABLE_WORKERS` is enabled.
  * Step execution respects dependency ordering.
  * Watcher enqueues full scan when changes exceed threshold; otherwise targeted jobs.
  * Job cancel endpoint sets cancel request/status behavior.
* Edge cases handled:
  * Deleted files trigger targeted `mark_deleted` jobs.
  * Missing/invalid snapshot paths are ignored safely.
* Validation logic:
  * Status/type constraints are enforced by DB checks and service logic.

## 6. User Flows

* Normal flow:
  * System starts -> startup scan queued -> steps run -> files/metadata become available.
* Error flow:
  * Step failures captured as `last_error`; job status reflects failed/partial states.
* Edge cases:
  * High file churn causes fallback to full-scan jobs and potential queue pressure.

## 7. API / Interfaces

* Endpoints:
  * `/api/v1/jobs*` as listed above.
* Input/output:
  * Query filters (`status`, `type`, `priority`) and paginated job lists; detailed job/step responses.
* Contracts:
  * Clients expect monotonic progress updates from step completion percentages.
* Internal interfaces:
  * `jobs.ServiceInterface`
  * `jobs.RepositoryInterface`
  * Worker scheduler/orchestrator interfaces.

## 8. Problems & Limitations

* Technical debt:
  * Legacy pipeline path still exists alongside orchestrated job model.
* Bugs or inconsistencies:
  * Queueing behavior between watcher and manual scan tasks can be hard to reason about under load.
* Performance issues:
  * Snapshot-based watcher (`filepath.WalkDir` every 5s) may be expensive on large trees.
* Missing validations:
  * No explicit global backpressure/rate controls on job creation.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscation detected.
* External code execution:
  * Worker ecosystem invokes Python metadata scripts and ffmpeg through dependent modules.
* Unsafe patterns:
  * Background jobs operate on filesystem paths and can perform recursive deletes/updates via downstream services.
* Injection risks:
  * Step payloads are JSON; strict validation of path payloads is critical.
* Hardcoded secrets:
  * None detected.
* Unsafe file/system access:
  * Continuous filesystem scanning and mutation increase blast radius if path guards fail.

## 10. Improvement Opportunities

* Refactors:
  * Retire legacy pipeline and keep one orchestration model.
* Architecture improvements:
  * Move from polling snapshot watcher to native FS event APIs with fallbacks.
* Scalability:
  * Add concurrency controls, queue quotas, and adaptive scheduling.
* UX improvements:
  * Expose detailed job timeline and retry controls in UI.

## 11. Acceptance Criteria

* Functional:
  * Startup and filesystem-change jobs are enqueued and processed end-to-end.
  * Jobs API reflects live status/progress and supports cancellation.
* Technical:
  * Dependency ordering is respected and persisted statuses are accurate.
  * Failed steps include diagnostic error details.
* Edge cases:
  * Massive change events degrade gracefully to full-scan mode.
  * Cancel requests stop further work for pending/runnable steps.

## 12. Open Questions

* Unknown behaviors:
  * SLO targets for scan completion latency by library size.
* Missing clarity in code:
  * No explicit dead-letter/retry policy documentation for persistent failures.
* Assumptions made:
  * Deployment has stable local disk and enough resources for recurring full-tree scans.
