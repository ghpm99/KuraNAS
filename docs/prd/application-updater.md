## 1. Feature Overview

* Name: Application Updater
* Summary: Checks latest GitHub release and applies in-place binary/assets update for Linux/Windows builds.
* Purpose: Enable upgrade workflow from running application without manual redeploy steps.
* Business value: Reduces operational friction and keeps deployments current.

## 2. Current Implementation

* How it works today: Updater queries GitHub releases API, selects OS-specific zip, downloads, validates size, extracts, replaces binary/resources, then schedules shutdown callback.
* Main flows:
  * User checks update status (`current vs latest`).
  * User triggers apply update.
  * Service downloads release asset to temp dir, verifies size, extracts with zip-slip protection.
  * Service replaces executable and selected directories (`dist`, `icons`, `translations`, `scripts`) while preserving `scripts/.venv`.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/update/status`
  * `POST /api/v1/update/apply`
* Key files involved (list with paths):
  * `backend/internal/api/v1/updater/handler.go`
  * `backend/internal/api/v1/updater/service.go`
  * `backend/internal/api/v1/updater/dto.go`
  * `frontend/src/service/update.ts`
  * `frontend/src/components/about/useAboutScreen.ts`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend about/update UX.
  * Backend updater handler/service with HTTP and filesystem operations.
* Data flow (step-by-step):
  * Status endpoint fetches latest release metadata from GitHub API.
  * Apply endpoint resolves matching asset name by `GOOS`.
  * File download uses separate HTTP client with larger timeout.
  * Zip extraction validates target paths to prevent zip-slip.
  * Binary and assets are copied into install directory; optional shutdown callback is scheduled.
* External integrations:
  * GitHub Releases API and direct asset download URLs.
* State management (if applicable):
  * Frontend query/mutation state only.

## 4. Data Model

* Entities involved:
  * Update status DTO and GitHub release/asset DTOs.
* Database tables / schemas:
  * None (stateless runtime operation).
* Relationships:
  * N/A.
* Important fields:
  * `current_version`, `latest_version`, `update_available`, `asset_name`, `asset_size`, `release_url`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Asset name is OS-dependent (`kuranas-linux.zip` or `kuranas-windows.zip`).
  * Apply requires matching asset and downloaded size equality check.
  * Zip extraction rejects entries escaping destination.
* Edge cases handled:
  * Missing matching asset returns explicit error.
  * Symlink resolution failures block update.
* Validation logic:
  * HTTP status checks and file size validation before extraction/apply.

## 6. User Flows

* Normal flow:
  * User opens About -> checks update status -> applies update -> app prepares replacement and restarts/shuts down.
* Error flow:
  * Network/API/download/extract/apply errors are returned and update is aborted.
* Edge cases:
  * Partial update failure can leave `.old` binary/temporary artifacts depending on failure point.

## 7. API / Interfaces

* Endpoints:
  * `GET /api/v1/update/status`
  * `POST /api/v1/update/apply`
* Input/output:
  * Status response includes release metadata; apply returns success/failure status.
* Contracts:
  * Frontend expects deterministic status fields and actionable error messages.
* Internal interfaces:
  * `updater.Service` methods `CheckForUpdate` and `DownloadAndApply`.

## 8. Problems & Limitations

* Technical debt:
  * Update operation is tightly coupled to packaging layout assumptions.
* Bugs or inconsistencies:
  * No transactional rollback mechanism for multi-step replacement.
* Performance issues:
  * Large downloads/extractions block operation and can timeout on weak networks/disks.
* Missing validations:
  * No cryptographic signature verification of downloaded assets.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * External download and executable replacement path is high-risk by design.
* External code execution:
  * Downloads and installs release package from internet source.
* Unsafe patterns:
  * Integrity check uses asset size only; size match is insufficient against tampering.
* Injection risks:
  * Update source URL is derived from GitHub API response; supply-chain trust is critical.
* Hardcoded secrets:
  * None detected.
* Unsafe file/system access:
  * Replaces running binary and recursively removes/copies asset directories (`os.RemoveAll` + copy operations).

## 10. Improvement Opportunities

* Refactors:
  * Separate download, verify, and apply phases with resumable state.
* Architecture improvements:
  * Add signed manifest/checksum verification (e.g., SHA-256 + signature).
* Scalability:
  * Support staged updates and bandwidth-efficient delta updates.
* UX improvements:
  * Add progress reporting and rollback guidance when update fails.

## 11. Acceptance Criteria

* Functional:
  * System can detect newer release and apply update for current OS.
* Technical:
  * Downloaded asset is validated and extracted safely without path traversal.
  * Binary and assets are replaced according to packaging conventions.
* Edge cases:
  * Missing asset/version mismatch/network failures return clear non-success responses and avoid partial corruption where possible.

## 12. Open Questions

* Unknown behaviors:
  * Expected restart orchestration in different deployment environments (service manager/container).
* Missing clarity in code:
  * No explicit rollback contract after failed binary replacement.
* Assumptions made:
  * GitHub release channel is the only trusted update source.
