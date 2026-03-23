## 1. Feature Overview

* Name: Analytics Overview
* Summary: Produces operational and content analytics across file library, processing pipeline, and activity trends.
* Purpose: Give users a high-level system and library health dashboard.
* Business value: Supports operational visibility, storage decisions, and engagement insights.

## 2. Current Implementation

* How it works today: `/analytics/overview` aggregates metrics from file, metadata, log, and worker-job data, with optional AI insights.
* Main flows:
  * Frontend requests overview by period (`24h`, `7d`, `30d`, `90d`).
  * Service executes repository aggregations for KPIs/timeseries/distributions/hot folders/duplicates.
  * Service computes filesystem capacity (OS-dependent implementation).
  * Optional AI summary is generated from collected metrics.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/analytics/overview?period=`
* Key files involved (list with paths):
  * `backend/internal/api/v1/analytics/handler.go`
  * `backend/internal/api/v1/analytics/service.go`
  * `backend/internal/api/v1/analytics/repository.go`
  * `backend/internal/api/v1/analytics/filesystem_stats_unix.go`
  * `backend/internal/api/v1/analytics/filesystem_stats_windows.go`
  * `frontend/src/service/analytics.ts`
  * `frontend/src/components/providers/analyticsProvider/index.tsx`
  * `frontend/src/pages/analytics/index.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend analytics provider/page.
  * Backend handler/service/repository + OS-specific capacity helpers.
* Data flow (step-by-step):
  * Request period is validated/clamped.
  * Repository runs aggregate SQL queries over `home_file`, metadata, jobs, and logs.
  * Service maps metrics into overview DTO sections.
  * Optional AI task turns numeric snapshot into textual insight.
* External integrations:
  * OS filesystem stat APIs.
  * Optional AI provider call for narrative insights.
* State management (if applicable):
  * React Query caches overview response by period.

## 4. Data Model

* Entities involved:
  * Analytics overview DTO, KPI series, distributions, queue/health summaries.
* Database tables / schemas:
  * `home_file`, `image_metadata`, `audio_metadata`, `video_metadata`
  * `worker_job`, `worker_step`
  * `log`
  * `recent_file` (recent activity slices)
* Relationships:
  * Analytics is computed across multiple tables without persistent materialized model.
* Important fields:
  * Storage sizes/counts, duplicate signatures, processing status counts, error counts.

## 5. Business Rules

* Explicit rules implemented in code:
  * Only predefined periods are accepted.
  * Overview merges file-library and pipeline health signals into one payload.
* Edge cases handled:
  * Empty datasets return zeroed metrics.
  * AI failure should not block base analytics response.
* Validation logic:
  * Period parsing and sane defaults.

## 6. User Flows

* Normal flow:
  * User opens Analytics -> selects period -> sees KPIs/charts/library coverage/hot folders.
* Error flow:
  * Aggregate query failure returns server error.
* Edge cases:
  * Sparse/new installations show low-information but valid zero states.

## 7. API / Interfaces

* Endpoints:
  * `GET /api/v1/analytics/overview`
* Input/output:
  * Query parameter `period`; response includes overview sections and optional insight text.
* Contracts:
  * Frontend expects stable object keys per section for chart rendering.
* Internal interfaces:
  * `analytics.ServiceInterface`
  * `analytics.RepositoryInterface`

## 8. Problems & Limitations

* Technical debt:
  * Heavy aggregation logic concentrated in one service/repository path.
* Bugs or inconsistencies:
  * Metric definitions are code-implicit; no formal analytics dictionary/versioning.
* Performance issues:
  * Expensive aggregate queries may slow response for very large libraries.
* Missing validations:
  * No explicit query timeout budget per analytics subsection.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code detected.
* External code execution:
  * None in analytics core.
* Unsafe patterns:
  * Optional AI narrative introduces external-data egress risk if enabled.
* Injection risks:
  * Period input is constrained; repository should remain strictly parameterized.
* Hardcoded secrets:
  * None in analytics module.
* Unsafe file/system access:
  * Reads filesystem capacity info only.

## 10. Improvement Opportunities

* Refactors:
  * Split overview computation into independent metric providers.
* Architecture improvements:
  * Add materialized views or periodic precomputation for high-cost metrics.
* Scalability:
  * Add background metric cache and TTL invalidation.
* UX improvements:
  * Add drill-down endpoints from summary cards to actionable detail pages.

## 11. Acceptance Criteria

* Functional:
  * Overview endpoint returns complete analytics payload for supported periods.
* Technical:
  * Empty library and large library both return valid structured responses.
  * AI insights remain optional and non-blocking.
* Edge cases:
  * Invalid period falls back or returns documented validation error.

## 12. Open Questions

* Unknown behaviors:
  * Target freshness SLA for analytics in relation to worker indexing lag.
* Missing clarity in code:
  * No explicit ownership of metric definitions across product/engineering.
* Assumptions made:
  * Analytics is informational and not used for billing or compliance decisions.
