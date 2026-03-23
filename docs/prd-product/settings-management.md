## 1. Feature Overview
- Name: Settings Management
- Summary: Persistent system preferences for library, indexing, players, appearance, and language selection.
- Target user: User configuring system behavior and UX defaults.
- Business value: Personalizes experience and controls operational defaults.

## 2. Problem Statement
- Users need persistent configuration instead of reapplying preferences every session.
- Without settings, behavior is static and not user-aligned.

## 3. Current Behavior
- Backend loads/saves a structured settings document (`system_preferences`) in `app_settings`.
- Frontend edits a local draft, tracks unsaved changes, and saves explicitly.
- Appearance settings are applied to CSS variables at runtime.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/settings`.
2. System loads saved settings.
3. User edits sections (library/indexing/players/appearance/language).
4. User saves changes; backend validates and persists normalized settings.

### 4.2 Alternative Flows
1. User resets draft to baseline settings loaded from backend.
2. User adjusts appearance and immediately sees updated visual tokens.

### 4.3 Error Scenarios
1. Invalid request payload returns bad request.
2. Save/load failures show snackbar or error state.

## 5. Functional Requirements
- The system must return current settings and available language options.
- The user can edit and persist all settings sections.
- The system should prevent save when no draft changes exist.

## 6. Business Rules
- Accent color allowed values: `violet`, `cyan`, `rose`.
- Slideshow interval allowed values: `4`, `8`, `12`, `20`.
- Watch paths are deduplicated and sanitized.
- Workers enabled flag in settings response reflects runtime env configuration.

## 7. Data Model (Business View)
- App Setting Document (`app_settings`): key-value persistence for configuration JSON.
- Settings groups: library, indexing, players, appearance, language.

## 8. Interfaces
- User interfaces: `/settings` page.
- APIs:
  - `GET /configuration/settings`
  - `PUT /configuration/settings`

## 9. Dependencies
- Configuration repository/service.
- Settings provider and snackbar feedback in frontend.

## 10. Limitations / Gaps
- No versioning or rollback for settings mutations.
- Runtime root path remains env-driven; watched paths do not automatically reconfigure scan scope.

## 11. Opportunities
- Add per-profile settings and audit trail.
- Add import/export settings templates.

## 12. Acceptance Criteria
- Given valid payload, when user saves settings, then settings are persisted and returned normalized.
- Given invalid accent/slideshow value, when user saves, then backend rejects request.
- Given unsaved changes are absent, when user is on settings page, then save action is disabled.

## 13. Assumptions
- Settings are globally scoped to deployment (single profile).
