## 1. Feature Overview
- Name: Application Updater
- Summary: Self-update capability that checks latest GitHub release and applies binary/assets package replacement.
- Target user: User maintaining app version without manual reinstall.
- Business value: Reduces maintenance friction and speeds version adoption.

## 2. Problem Statement
- Manual upgrades are error-prone and inconvenient.
- Without updater automation, users delay critical updates.

## 3. Current Behavior
- Checks `releases/latest` from GitHub API and compares semantic version against current build version.
- Selects OS-specific asset (`kuranas-linux.zip` or `kuranas-windows.zip`).
- Downloads zip, validates expected file size, extracts with zip-slip protection, replaces binary/assets.
- Preserves `scripts/.venv` during scripts directory replacement.
- Schedules app shutdown callback after successful apply.

## 4. User Flows
### 4.1 Main Flow
1. User opens About update panel.
2. UI requests `/update/status` and shows current/latest metadata.
3. If update available, user triggers apply.
4. Backend downloads, validates, extracts, applies update, and responds success.

### 4.2 Alternative Flows
1. User checks update status without applying.
2. No update available path shows up-to-date state.

### 4.3 Error Scenarios
1. GitHub API/network failure returns status/apply error.
2. Matching asset not found for platform -> apply fails.
3. Download size mismatch -> apply aborted.
4. Extract/apply failure -> update aborted and temp files cleaned.

## 5. Functional Requirements
- The system must report update status including versions and release metadata.
- The user can trigger update download and apply.
- The system must safely extract and apply package contents.

## 6. Business Rules
- Version compare uses semantic major/minor/patch parser; non-numeric versions fall back to `0.0.0`.
- Apply operation validates downloaded asset size against release metadata.
- Platform asset selection is OS-dependent.

## 7. Data Model (Business View)
- No persistent update table; transient DTOs:
  - `UpdateStatusDto` (current/latest/version flags/release notes/asset info)
  - GitHub release and asset payload mappings.

## 8. Interfaces
- User interfaces: About update panel.
- APIs:
  - `GET /update/status`
  - `POST /update/apply`

## 9. Dependencies
- GitHub Releases API.
- Filesystem write permissions on installation directory.
- Logger and optional shutdown callback integration.

## 10. Limitations / Gaps
- No staged rollout/canary or rollback strategy in current flow.
- Update source is fixed to one repository endpoint.

## 11. Opportunities
- Add signed artifact verification.
- Add rollback checkpoint and post-update health verification.

## 12. Acceptance Criteria
- Given newer release exists, when status endpoint is called, then `update_available=true` and release metadata is returned.
- Given apply is triggered with valid asset, when process completes, then binary/assets are replaced and success response is returned.
- Given asset size mismatch, when apply runs, then update is aborted with error.

## 13. Assumptions
- Runtime has network access to GitHub and permission to modify install files.
