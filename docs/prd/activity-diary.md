## 1. Feature Overview

* Name: Activity Diary
* Summary: Tracks user activities with start/end timestamps, notes, duplication, and summary metrics.
* Purpose: Provide lightweight activity journaling inside the system.
* Business value: Adds personal productivity tracking and sticky daily usage behavior.

## 2. Current Implementation

* How it works today: Diary endpoints allow create/read/update/duplicate plus summary aggregation.
* Main flows:
  * User creates activity entry.
  * Service auto-closes previously active entry before creating new one.
  * User can update or duplicate entries.
  * Summary endpoint aggregates counts/duration/longest activity.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/diary/`
  * `GET /api/v1/diary/summary`
  * `POST /api/v1/diary/`
  * `PUT /api/v1/diary/:id`
  * `POST /api/v1/diary/copy`
* Key files involved (list with paths):
  * `backend/internal/api/v1/diary/handler.go`
  * `backend/internal/api/v1/diary/service.go`
  * `backend/internal/api/v1/diary/repository.go`
  * `backend/internal/api/v1/diary/dto.go`
  * `frontend/src/service/activityDiary.ts`
  * `frontend/src/components/providers/activityDiaryProvider/index.tsx`
  * `frontend/src/pages/activityDiary/index.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend provider/service + diary page.
  * Backend handler -> service -> repository.
* Data flow (step-by-step):
  * Create request parses diary payload.
  * Service finds active activity and sets its end time if needed.
  * Repository inserts new diary row.
  * List/summary endpoints query and map rows to response DTOs.
* External integrations:
  * None.
* State management (if applicable):
  * React context/provider maintains diary list and mutation state.

## 4. Data Model

* Entities involved:
  * Activity diary entry.
* Database tables / schemas:
  * `activity_diary(id, name, description, start_time, end_time)`.
* Relationships:
  * Standalone table, no explicit foreign keys.
* Important fields:
  * `start_time`, `end_time` for duration semantics.

## 5. Business Rules

* Explicit rules implemented in code:
  * Creating a new diary entry closes existing open entry.
  * Summary computes total duration and longest item based on selected range.
* Edge cases handled:
  * Empty diary list returns empty/zero summary values.
* Validation logic:
  * DTO parsing for required fields and ids.

## 6. User Flows

* Normal flow:
  * User creates activity -> entry starts -> user ends/updates later -> summary reflects cumulative data.
* Error flow:
  * Invalid IDs or malformed payloads return errors.
* Edge cases:
  * Duplicate endpoint depends on request field naming consistency.

## 7. API / Interfaces

* Endpoints:
  * `/api/v1/diary/*` routes listed above.
* Input/output:
  * Diary create/update/copy payloads and paginated/list/summary responses.
* Contracts:
  * Frontend expects consistent id field naming and date format.
* Internal interfaces:
  * `diary.ServiceInterface`
  * `diary.RepositoryInterface`

## 8. Problems & Limitations

* Technical debt:
  * Feature has minimal domain abstractions and limited reporting options.
* Bugs or inconsistencies:
  * Duplicate contract mismatch risk: frontend sends `ID` while backend expects JSON `id` (`DiaryId` field in DTO).
* Performance issues:
  * None significant at current scale.
* Missing validations:
  * No rule enforcing `end_time >= start_time` on updates.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated behavior detected.
* External code execution:
  * None.
* Unsafe patterns:
  * No auth boundary means diary data is not user-isolated.
* Injection risks:
  * Depends on repository query parameterization; no dynamic execution observed.
* Hardcoded secrets:
  * None detected.
* Unsafe file/system access:
  * None.

## 10. Improvement Opportunities

* Refactors:
  * Add dedicated domain rules object for activity lifecycle constraints.
* Architecture improvements:
  * Add authenticated owner scoping for diary entries.
* Scalability:
  * Add indexed time-range queries and optional archival.
* UX improvements:
  * Add filters (date range, tags, status) and richer summary charts.

## 11. Acceptance Criteria

* Functional:
  * Users can create, edit, duplicate, and view diary entries and summaries.
* Technical:
  * Creating a new entry closes previous active one deterministically.
  * Duplicate endpoint accepts documented payload and returns copied entry.
* Edge cases:
  * Invalid IDs/payloads return clear 4xx responses.
  * Summary endpoints return valid zero states with no data.

## 12. Open Questions

* Unknown behaviors:
  * Whether summaries should be timezone-aware per user preference.
* Missing clarity in code:
  * No explicit API contract doc for duplicate payload naming.
* Assumptions made:
  * Diary is currently single-tenant/session-oriented.
