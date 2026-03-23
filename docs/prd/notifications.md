## 1. Feature Overview

* Name: Notifications
* Summary: Delivers in-app notification feed with unread counts, grouped events, and mark-read operations.
* Purpose: Surface operational status and user-relevant events from background processing.
* Business value: Improves transparency and trust by making asynchronous system behavior visible.

## 2. Current Implementation

* How it works today: Worker and services create notification records; API supports listing, detail retrieval, and read state updates.
* Main flows:
  * System emits notification (`GroupOrCreate`) for recurring event classes.
  * User opens notifications page and fetches list/count.
  * User marks one/all notifications as read.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/notifications`
  * `GET /api/v1/notifications/unread-count`
  * `GET /api/v1/notifications/:id`
  * `PUT /api/v1/notifications/:id/read`
  * `PUT /api/v1/notifications/read-all`
* Key files involved (list with paths):
  * `backend/internal/api/v1/notifications/handler.go`
  * `backend/internal/api/v1/notifications/service.go`
  * `backend/internal/api/v1/notifications/repository.go`
  * `backend/internal/worker/worker.go`
  * `frontend/src/service/notifications.ts`
  * `frontend/src/components/providers/notificationProvider/index.tsx`
  * `frontend/src/pages/notifications/index.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend notification provider/page.
  * Backend handler/service/repository and worker emitter integration.
* Data flow (step-by-step):
  * Event trigger calls service create/group APIs.
  * Repository persists/updates notification row(s).
  * List endpoint returns sorted notifications.
  * Read endpoints update `is_read` flags.
* External integrations:
  * None.
* State management (if applicable):
  * React context/provider manages unread count and list refresh.

## 4. Data Model

* Entities involved:
  * Notification record and grouped-notification metadata.
* Database tables / schemas:
  * `notifications`
* Relationships:
  * Standalone table; grouping via `group_key`.
* Important fields:
  * `type`, `title`, `message`, `is_read`, `group_key`, `group_count`, `is_grouped`, `metadata`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Grouping behavior aggregates same-key events within logic windows.
  * Error notifications are not grouped.
  * Read-all marks unread set in bulk.
* Edge cases handled:
  * Missing notification id returns not found.
* Validation logic:
  * DTO parsing for IDs and paging fields.

## 6. User Flows

* Normal flow:
  * Background event occurs -> notification appears -> user reads/marks read.
* Error flow:
  * Invalid ID on read/detail returns 4xx/5xx.
* Edge cases:
  * High event volume relies on grouping to limit feed noise.

## 7. API / Interfaces

* Endpoints:
  * `/api/v1/notifications*`
* Input/output:
  * List query params, unread count response, read mutation responses.
* Contracts:
  * Frontend expects grouped notification shape and count fields.
* Internal interfaces:
  * `notifications.ServiceInterface`
  * `notifications.RepositoryInterface`

## 8. Problems & Limitations

* Technical debt:
  * Notification categories/types are string-based and loosely governed.
* Bugs or inconsistencies:
  * Grouping policy is code-defined with limited configurability.
* Performance issues:
  * Cleanup runs hourly; very high notification rates may still grow table quickly.
* Missing validations:
  * No explicit max message size controls at service boundary.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code found.
* External code execution:
  * None.
* Unsafe patterns:
  * Notification messages can contain dynamic error text; frontend rendering must avoid unsafe HTML insertion.
* Injection risks:
  * Metadata JSON content should be treated as untrusted payload.
* Hardcoded secrets:
  * None detected.
* Unsafe file/system access:
  * None direct.

## 10. Improvement Opportunities

* Refactors:
  * Centralize notification type registry and schema validation.
* Architecture improvements:
  * Add event bus abstraction for producers and channel-specific consumers.
* Scalability:
  * Add retention policy controls and partitioning for high-volume installations.
* UX improvements:
  * Add filters by type/source and actionable deep links.

## 11. Acceptance Criteria

* Functional:
  * Notifications can be created, listed, counted, read individually, and read in bulk.
* Technical:
  * Grouped notifications increment counts correctly and preserve latest timestamps.
* Edge cases:
  * Invalid IDs return clear errors.
  * Cleanup process removes old notifications per configured retention policy.

## 12. Open Questions

* Unknown behaviors:
  * Desired user-configurable retention and mute preferences.
* Missing clarity in code:
  * No explicit SLA for notification delivery latency from event to UI.
* Assumptions made:
  * Notification channel is in-app only (no push/email/webhook channels).
