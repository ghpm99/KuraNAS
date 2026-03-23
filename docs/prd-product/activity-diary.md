## 1. Feature Overview
- Name: Activity Diary
- Summary: Time-based manual activity tracking with active entry continuity, summary metrics, and entry duplication.
- Target user: User logging ongoing tasks/work sessions.
- Business value: Provides lightweight personal productivity logging inside NAS workspace.

## 2. Problem Statement
- Users need quick logging of what they are doing and for how long.
- Without diary tracking, activity history and time spend are not visible.

## 3. Current Behavior
- Creating a new diary entry auto-closes previous active entry by setting previous `end_time` to new `start_time`.
- Summary endpoint reports recent window stats (last hour): total activities, total duration, longest activity.
- Frontend validates input (length, character pattern, non-empty) before create.
- Entry duplication creates a new active entry from selected previous entry.

## 4. User Flows
### 4.1 Main Flow
1. User opens activity diary page.
2. User enters name/description and submits.
3. Backend creates new entry and closes prior open entry.
4. UI refreshes list and summary.

### 4.2 Alternative Flows
1. User duplicates existing activity to start same activity quickly.
2. User monitors live duration using frontend time ticker.

### 4.3 Error Scenarios
1. Invalid form input (client validation) blocks submission.
2. Invalid payload/server error returns create/duplicate failure feedback.
3. Duplicate by missing ID results in backend error.

## 5. Functional Requirements
- The user can create diary entries with name and optional description.
- The system must maintain a single active interval by closing previous entry on create.
- The user can duplicate an entry.
- The system should provide summary and list retrieval endpoints.

## 6. Business Rules
- New entry `start_time` is current server time.
- Prior latest entry gets `end_time` set to new start time.
- Summary period is fixed to last 1 hour.
- Ongoing entries (no end_time) compute duration using current time.

## 7. Data Model (Business View)
- Activity Diary (`activity_diary`): name, description, start_time, end_time.
- Derived fields in DTO: duration seconds/formatted, in_progress.

## 8. Interfaces
- User interfaces: `/internal/activity-diary` with form, list, summary, actions.
- APIs:
  - `GET /diary/`
  - `GET /diary/summary`
  - `POST /diary/`
  - `POST /diary/copy`

## 9. Dependencies
- Logger service for handler operations.
- i18n messages for frontend feedback and backend responses.

## 10. Limitations / Gaps
- Update endpoint shape is inconsistent with other APIs and appears legacy.
- No explicit delete/archive workflow for diary entries.

## 11. Opportunities
- Add date-range filtering and exports.
- Add tags/categories and richer reporting.

## 12. Acceptance Criteria
- Given an active entry exists, when user creates a new entry, then previous entry is closed and new entry starts now.
- Given valid diary form, when submitted, then entry appears in list and summary refreshes.
- Given duplicate action with valid ID, when executed, then a new active entry is created from source activity.

## 13. Assumptions
- Diary is single-user scoped by deployment context.
