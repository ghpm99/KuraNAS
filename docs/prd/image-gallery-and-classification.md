## 1. Feature Overview

* Name: Image Gallery and Classification
* Summary: Presents image-focused views (recent, captures, photos, folders/albums-like groupings) and classifies image content metadata.
* Purpose: Improve discoverability of photo/image assets beyond generic file browsing.
* Business value: Increases engagement with visual media and supports smarter retrieval/categorization.

## 2. Current Implementation

* How it works today: Frontend images page consumes `files` image endpoints; backend returns filtered image lists plus metadata/classification fields.
* Main flows:
  * Frontend requests image lists via `getImageFiles` and section-specific filters.
  * Backend retrieves image files from `home_file` + `image_metadata` and sorts (e.g., by date).
  * Classification data can be assigned by heuristic/AI-assisted backend logic.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/files/images`
  * `GET /api/v1/files/thumbnail/:id`
  * Worker/processing path that enriches `image_metadata` and classification.
* Key files involved (list with paths):
  * `backend/internal/api/v1/files/image_classification.go`
  * `backend/internal/api/v1/files/service.go`
  * `backend/internal/api/v1/files/metadata_repository.go`
  * `frontend/src/pages/images/index.tsx`
  * `frontend/src/components/providers/imageProvider/imageProvider.tsx`
  * `frontend/src/service/files.ts`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend image provider/page + query hooks.
  * Backend files service + metadata repository + optional AI service usage.
* Data flow (step-by-step):
  * Indexed images are stored in `home_file`.
  * Metadata extraction populates `image_metadata`.
  * Classification computes `classification_category` and confidence.
  * API returns image DTOs to image provider.
  * UI maps data into image sections and preview cards.
* External integrations:
  * Python metadata scripts and optional AI classification provider.
* State management (if applicable):
  * React context in `imageProvider` with query-driven refresh.

## 4. Data Model

* Entities involved:
  * Image file and image metadata record.
* Database tables / schemas:
  * `home_file`
  * `image_metadata`
* Relationships:
  * `image_metadata.file_id -> home_file.id`.
* Important fields:
  * `width`, `height`, EXIF/date fields.
  * `classification_category`, `classification_confidence`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Only image-type files are returned for image-specific APIs.
  * Classification pipeline uses heuristics first and AI fallback for low-confidence outcomes.
* Edge cases handled:
  * Images without metadata still appear with degraded feature set.
  * Unknown classification defaults to generic categories.
* Validation logic:
  * File type checks and metadata parse guards.

## 6. User Flows

* Normal flow:
  * User opens Images -> navigates image sections -> loads thumbnails -> opens file in viewer/path.
* Error flow:
  * Thumbnail/metadata failures show fallback visuals and missing metadata states.
* Edge cases:
  * Corrupted image metadata or unsupported formats appear without enriched fields.

## 7. API / Interfaces

* Endpoints:
  * `GET /api/v1/files/images`
  * `GET /api/v1/files/thumbnail/:id`
* Input/output:
  * Pagination/sort query parameters and image DTO list responses.
* Contracts:
  * Frontend expects optional metadata/classification fields and resilient null handling.
* Internal interfaces:
  * Files service/repository metadata access APIs used by image provider paths.

## 8. Problems & Limitations

* Technical debt:
  * Image feature is embedded inside broad files module, limiting dedicated image-domain evolution.
* Bugs or inconsistencies:
  * Section semantics (albums/folders/captures) are partially frontend-derived rather than explicit backend contracts.
* Performance issues:
  * Thumbnail fetch bursts can cause IO spikes.
* Missing validations:
  * No explicit quality threshold governance for AI-derived classification labels.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code detected.
* External code execution:
  * Metadata extraction path uses Python script execution (`RunPythonScript`).
* Unsafe patterns:
  * AI classification relies on external provider output; output sanitization for labels should remain strict.
* Injection risks:
  * Limited direct injection exposure, but metadata text should be treated as untrusted content in UI rendering.
* Hardcoded secrets:
  * None found in image-specific module.
* Unsafe file/system access:
  * Reads and thumbnail generation depend on filesystem paths from indexed records.

## 10. Improvement Opportunities

* Refactors:
  * Isolate image-domain service/API from generic files service.
* Architecture improvements:
  * Add deterministic local classifier abstraction with explicit confidence calibration.
* Scalability:
  * Pre-generate/cache popular thumbnails and classify incrementally.
* UX improvements:
  * Add explicit filters by date, resolution, orientation, and category confidence.

## 11. Acceptance Criteria

* Functional:
  * Users can browse image-specific sections and view thumbnails reliably.
* Technical:
  * Classification fields are persisted and returned consistently when available.
  * Missing metadata does not break API contract.
* Edge cases:
  * Corrupt or unsupported images are handled without crashing list endpoints.

## 12. Open Questions

* Unknown behaviors:
  * Exact taxonomy governance for image `classification_category`.
* Missing clarity in code:
  * No explicit retention/recompute policy for classification when models change.
* Assumptions made:
  * Images page relies on files-domain APIs rather than a dedicated image API boundary.
