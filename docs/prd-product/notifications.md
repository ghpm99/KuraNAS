## 1. Feature Overview
- Name: Notifications
- Summary: In-app notification center with unread counters, filtering, read-state actions, and grouping of repeated events.
- Target user: User monitoring background activity and system events.
- Business value: Improves visibility of asynchronous operations and issues.

## 2. Problem Statement
- Background processing and scans need user-visible status communication.
- Without notifications, users lose operational awareness.

## 3. Current Behavior
- Notification bell shows unread count and quick dropdown list.
- Full page (`/notifications`) supports tab filtering (all/unread/type).
- Backend supports mark one/all as read and unread count endpoint.
- Grouping behavior:
  - grouped by `group_key + type` inside a short window.
  - errors are never grouped.

## 4. User Flows
### 4.1 Main Flow
1. Worker or service emits notification.
2. User sees unread badge in header bell.
3. User opens dropdown/page, reads notification, and marks as read.

### 4.2 Alternative Flows
1. User marks all notifications as read in dropdown/page.
2. User filters notifications by unread or severity type.

### 4.3 Error Scenarios
1. Invalid notification ID -> bad request.
2. Missing notification -> not found.
3. Notification service failure -> list/count/read operations fail.

## 5. Functional Requirements
- The system must store and list notifications with pagination.
- The user can filter notifications by type and read state.
- The user can mark one notification or all notifications as read.
- The system should group repetitive non-error notifications.

## 6. Business Rules
- Allowed types: info, success, warning, error, system.
- Group window is 60 seconds for groupable notifications.
- Grouped message updates as count grows.
- Cleanup job deletes old notifications periodically (worker cleanup loop).

## 7. Data Model (Business View)
- Notification (`notifications`): type, title, message, metadata, is_read, group_key, group_count, is_grouped, created_at.

## 8. Interfaces
- User interfaces: header bell/dropdown, `/notifications` page.
- APIs:
  - `GET /notifications`
  - `GET /notifications/unread-count`
  - `GET /notifications/:id`
  - `PUT /notifications/:id/read`
  - `PUT /notifications/read-all`

## 9. Dependencies
- Worker/event producers emitting notifications.
- Notification provider polling unread count.

## 10. Limitations / Gaps
- No notification preference management per type.
- No explicit link/action payload standard for actionable notifications.

## 11. Opportunities
- Add notification deep links to related resources.
- Add mute/snooze categories and retention controls.

## 12. Acceptance Criteria
- Given unread notifications exist, when header renders, then unread badge count is shown.
- Given user marks notification as read, when list refreshes, then item read-state is updated and unread count decreases.
- Given repeated grouped event within grouping window, when inserted, then existing notification count/message are incremented.

## 13. Assumptions
- Notification volume is moderate and pagination size defaults are sufficient for current use.
