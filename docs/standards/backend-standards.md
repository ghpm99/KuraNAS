# Backend Standards (Canonical)

This document is the canonical source for backend implementation patterns in this repository.

## Agent Enforcement
- Before any backend change, read this file first.
- If a rule here conflicts with local file style, prefer this file and then normalize surrounding code when safe.
- If a requested change conflicts with these rules, explicitly call out the conflict before implementing.

## 1) Architecture
- Stack is `Go + Gin + database/sql` with startup in `backend/cmd/nas/main.go`.
- Keep HTTP modules under `backend/internal/api/v1/<feature>`.
- Keep backend composition/wiring in `backend/internal/app/*` (context creation, route registration, app lifecycle).
- Keep shared infrastructure under `backend/pkg/*` (database, i18n, logger, utils, media helpers).
- Follow Clean Code and SRP: handlers, services, repositories, and workers must have focused responsibilities.

## 2) Layering and Responsibilities
- Enforce layered flow: `Handler -> Service -> Repository`.
- Handlers must parse request input, call service, map errors/status codes, and return JSON/files.
- Services must contain business rules, orchestration, and transaction boundaries.
- Repositories must contain persistence logic only (SQL/query execution and model mapping).
- Do not bypass layers (for example, handler querying database directly).

## 3) HTTP/API Rules
- Register routes centrally in `backend/internal/app/routes.go`.
- Group endpoints by domain (`/files`, `/music`, `/video`, `/analytics`, `/diary`, `/configuration`, `/update`).
- Use `utils.ParseInt` and similar helpers for input parsing/validation where applicable.
- Use consistent HTTP status codes (`400` bad request, `404` not found, `500` internal errors, `200/201` success).
- Return structured JSON responses and avoid leaking internal implementation details.

## 4) Service Layer Rules
- Business logic belongs in services (`service.go`) with explicit interfaces.
- Complex write operations must run inside transactions (`withTransaction` + `DbContext.ExecTx`).
- Keep services deterministic and testable by depending on interfaces, not concrete DB primitives.
- Keep DTO/model transformation logic explicit and close to domain types.
- Background task enqueueing (worker tasks) must be controlled by service/application orchestration, not by handlers directly.

## 5) Repository and SQL Rules
- Repositories must use SQL definitions from `backend/pkg/database/queries/<feature>`.
- Add/modify SQL in `.sql` files first and expose them through feature query loaders (`*.go` in queries folders).
- Keep repository methods narrow, predictable, and named by intent (`Get`, `Create`, `Update`, `Delete`, `Upsert`).
- Wrap DB errors with context (`fmt.Errorf(...: %w, err)`) and avoid silent failures.
- Preserve pagination conventions using `utils.PaginationResponse` and `UpdatePagination()`.

## 6) Database and Migrations
- Database configuration/bootstrap must stay centralized in `backend/pkg/database`.
- Schema changes must be added as new migration files in `backend/pkg/database/migrations/queries`.
- Register new migrations in `backend/pkg/database/migrations/migrations.go`.
- Migrations must be idempotent, ordered, and safe to run on startup.
- Never edit old migration behavior in a way that breaks existing installations; create a new migration instead.

## 7) Workers and Async Processing
- Worker orchestration remains in `backend/internal/worker/*`.
- Worker startup must respect `ENABLE_WORKERS` config flag.
- Long-running/background file scan, checksum, thumbnails, and video playlist generation must run through worker tasks.
- Avoid placing heavy filesystem processing inside HTTP request handlers.
- Ensure worker logic logs meaningful lifecycle and error states.

## 8) i18n and User-Facing Messages (Mandatory)
- Never hardcode user-facing messages that should be translatable.
- Backend translations must come from JSON files loaded by `backend/pkg/i18n` (`LoadTranslations`, `GetMessage`, `Translate`).
- Any new user-visible key must be added to translation JSON files under `backend/translations`.
- Keep key naming stable and semantic to preserve frontend/backend consistency.

## 9) Logging and Error Handling
- Use project logging service (`backend/pkg/logger`) for request/operation lifecycle when applicable.
- Track pending/success/error states for relevant operations, including API handlers with request context (IP, action, status).
- Prefer explicit error propagation over panic; panic is acceptable only for unrecoverable startup/bootstrap failures.
- Do not swallow errors; return actionable context to upper layers.

## 10) Configuration and Environment
- Runtime configuration must be centralized in `backend/internal/config`.
- Avoid hardcoded environment-specific values in feature code.
- New configuration flags must be added to config struct and consumed through config package access.
- Keep CORS and server-level concerns in app/router bootstrap, not scattered across handlers.
- Backend must remain buildable for both Windows and Linux targets.
- OS-specific syscalls/configuration must be isolated in dedicated files using Go build tags for each target OS.
- Prefer extending/reusing existing OS-specific files when context fits; create new files only when there is no suitable existing location.

## 11) Testing
- Use Go `testing` with tests colocated as `*_test.go` and additional suites under `backend/tests`.
- Prefer interface-based unit tests for handlers/services/repositories and keep integration tests explicit.
- For changed backend features, run at least:
- `cd backend && go test ./... -cover`
- `make -C backend test`
- Keep new business logic covered by targeted tests (success, validation errors, persistence errors).

## 12) Pull Request Checklist (Backend)
- API behavior change documented and reflected in handlers/service/repository layers.
- SQL or persistence changes include new migration/query updates when needed.
- i18n keys/messages updated for any new user-visible backend message.
- Worker/task impact assessed when filesystem/media behavior changes.
- Tests updated and executed for affected modules.
- If a new backend pattern is introduced, update this standards file in the same PR.

## Change Log
- 2026-03-05: Initial canonical backend standards file created based on current backend implementation.
- 2026-03-05: Added cross-platform rule (Windows/Linux) and mandatory build-tag isolation for OS-specific code.
