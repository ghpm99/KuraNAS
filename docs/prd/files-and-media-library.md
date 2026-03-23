## 1. Feature Overview

* Name: Files and Media Library
* Summary: Core filesystem indexing and browsing feature for files/folders, plus upload/download/stream/media metadata and thumbnail retrieval.
* Purpose: Provide NAS-style file management across documents, images, audio, and video.
* Business value: Primary product capability that enables users to organize and consume stored content from a single interface.

## 2. Current Implementation

* How it works today: Files are indexed in `home_file`; API serves tree/list/search-like filters, metadata reports, binary access, and file operations.
* Main flows:
  * UI requests file tree or path listing with pagination/category.
  * API reads indexed records, converts absolute paths to relative outputs.
  * Operations endpoints mutate filesystem and schedule/index updates.
  * Upload endpoint stores file then enqueues worker job steps.
  * Thumbnail/blob/stream endpoints serve media bytes for playback and previews.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/files/`, `/tree`, `/:id`, `/path`, `/recent`, `/thumbnail/:id`, `/video-thumbnail/:id`, `/video-preview/:id`, `/blob/:id`, `/stream/:id`, `/video-stream/:id`
  * `POST /api/v1/files/update`, `/upload`, `/folder`, `/move`, `/copy`, `/rename`, `/starred/:id`
  * `DELETE /api/v1/files/path`
  * Reports: `/total-space-used`, `/total-files`, `/total-directory`, `/report-size-by-format`, `/top-files-by-size`, `/duplicate-files`, `/images`, `/music`, `/videos`
* Key files involved (list with paths):
  * `backend/internal/api/v1/files/handler.go`
  * `backend/internal/api/v1/files/service.go`
  * `backend/internal/api/v1/files/repository.go`
  * `backend/internal/api/v1/files/operations_service.go`
  * `backend/internal/api/v1/files/recent_service.go`
  * `backend/internal/api/v1/files/metadata_repository.go`
  * `backend/internal/api/v1/files/image_classification.go`
  * `frontend/src/service/files.ts`
  * `frontend/src/components/providers/fileProvider/index.tsx`
  * `frontend/src/pages/files/index.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend page/provider calls typed service methods.
  * Backend handler -> service -> repository; operation-specific service for FS mutations.
* Data flow (step-by-step):
  * Request enters Gin route and parses query params.
  * Service builds filter/pagination and calls SQL repository.
  * Path normalization converts between relative and configured entrypoint absolute paths.
  * For writes, operations service resolves target path inside entrypoint and executes OS file operations.
  * Upload flow writes multipart files, then creates worker job with dependent steps (`persist`, `metadata`, `checksum`, optional thumbnail/playlist indexing).
* External integrations:
  * Local filesystem, ffmpeg process for video artifacts, HTTP Range streaming behavior.
* State management (if applicable):
  * React Query cache + FileProvider local selection/navigation state.

## 4. Data Model

* Entities involved:
  * File (`home_file`), recent file access, image/audio/video metadata.
* Database tables / schemas:
  * `home_file`
  * `recent_file`
  * `image_metadata`
  * `audio_metadata`
  * `video_metadata`
* Relationships:
  * Metadata tables reference `home_file.id`.
  * `recent_file.file_id` references files by id.
* Important fields:
  * `home_file.path`, `parent_path`, `type`, `checksum`, `deleted_at`, `starred`.
  * `image_metadata.classification_category`, `classification_confidence`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Category filter supports `all`, `recent`, `starred`.
  * File operations must resolve under configured entrypoint.
  * Stream endpoints support range requests for media.
  * Upload creates job steps conditionally by file type.
* Edge cases handled:
  * Missing file returns not found.
  * Deleted/missing physical file checks before blob/stream.
  * Duplicate detection via checksum/size reporting queries.
* Validation logic:
  * Request param parsing and path checks in service/operation layer.
  * File/folder creation and rename conflict checks.

## 6. User Flows

* Normal flow:
  * User opens Files/Favorites -> browses tree -> previews/downloads/streams -> performs move/copy/rename/delete/upload actions.
* Error flow:
  * Invalid paths or non-existing targets return error responses.
  * Stream/blob request for unsupported state returns 4xx/5xx with message.
* Edge cases:
  * Large folder updates trigger background jobs; UI may display stale state until reindex completes.

## 7. API / Interfaces

* Endpoints:
  * Full `/api/v1/files/*` suite listed in section 2.
* Input/output:
  * Query params for pagination/category/path and JSON/form payloads for operations.
  * Responses include paginated file DTOs, stats/reports, binary streams.
* Contracts:
  * Frontend expects relative paths and typed pagination envelope.
* Internal interfaces:
  * `files.ServiceInterface`
  * `files.RepositoryInterface`
  * `files.MetadataRepositoryInterface`
  * `files.RecentFileServiceInterface`

## 8. Problems & Limitations

* Technical debt:
  * Files domain is very broad (browse + analytics + media + operations), increasing coupling.
* Bugs or inconsistencies:
  * Some operation/report logic is duplicated across service/repository helpers.
* Performance issues:
  * Heavy list/report queries can be expensive on very large libraries.
  * Thumbnail generation on demand can increase latency.
* Missing validations:
  * No explicit per-user authorization boundaries; behavior is IP/session-agnostic.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code found.
* External code execution:
  * Uses `exec.Command("ffmpeg", ...)` for video thumbnail/preview generation.
* Unsafe patterns:
  * File mutation endpoints can delete/move/copy directories recursively; strict operational controls are required.
* Injection risks:
  * Path handling is guarded by entrypoint checks, but path canonicalization should be continuously fuzz-tested.
* Hardcoded secrets:
  * None detected in this module.
* Unsafe file/system access:
  * Uses `os.RemoveAll` for deletion; risk is high if path checks regress.

## 10. Improvement Opportunities

* Refactors:
  * Split files module into browse, operations, and media-delivery bounded contexts.
* Architecture improvements:
  * Introduce policy/authorization middleware before mutating endpoints.
* Scalability:
  * Add indexed/full-text strategy for name/path queries and async report materialization.
* UX improvements:
  * Add operation progress/events and optimistic UI rollback.

## 11. Acceptance Criteria

* Functional:
  * Users can browse, upload, organize, and stream files from the configured entrypoint.
* Technical:
  * All operations are constrained to entrypoint and persisted index remains consistent after jobs complete.
  * Report/stat endpoints return stable typed outputs.
* Edge cases:
  * Invalid/missing paths return clear non-200 responses.
  * Large uploads still enqueue processing jobs without blocking request lifecycle.

## 12. Open Questions

* Unknown behaviors:
  * Expected consistency model between watcher-driven updates and manual operations.
* Missing clarity in code:
  * No explicit rate limits for expensive report and streaming endpoints.
* Assumptions made:
  * System is intended for trusted private-network usage without multi-tenant isolation.
