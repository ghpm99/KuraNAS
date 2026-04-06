# Repository Guidelines

## Project Overview

KuraNAS is a personal NAS (Network Attached Storage) system with a Go backend and React/TypeScript frontend. The backend serves the frontend as a SPA and exposes a REST API at `/api/v1`.

## Project Structure & Module Organization

`backend/` contains the Go API and worker pipeline. Main entrypoint is `backend/cmd/nas/main.go`; HTTP modules live under `backend/internal/api/v1`; shared infrastructure is in `backend/pkg` (database, i18n, utils, media helpers). SQL migrations and query files are in `backend/pkg/database/migrations` and `backend/pkg/database/queries`.

`frontend/` is a Vite + React + TypeScript app. Core UI code is in `frontend/src/components`, route-level pages in `frontend/src/pages`, API clients in `frontend/src/service`, and shared types/utilities in `frontend/src/types` and `frontend/src/utils`.

`mobile/` is a native Android application targeting the **Samsung Galaxy Tab 2 7.0 (GT-P3110)** running **Android 4.1.2 (API level 16)**. The stack must be **Java + XML Views + AppCompat** — do not use Kotlin or Jetpack Compose, as neither is appropriate for this device (Compose requires API 21+; Kotlin adds runtime overhead and stdlib compatibility risks on Dalvik/API 16). All mobile development decisions — API usage, UI layout, library compatibility, and feature support — must account for this specific device and OS version constraint. The module structure follows standard Android Gradle conventions: `mobile/app/` contains the application module, `mobile/build.gradle` holds top-level build configuration.

`build/`, `frontend/dist/`, and `frontend/coverage/` are generated artifacts.

## Architecture

### Backend (`backend/`)

**Module**: `nas-go/api`
**Framework**: Gin (HTTP router), PostgreSQL (production database)
**Entry point**: `cmd/nas/main.go`

Layered architecture per domain: **Repository → Service → Handler**

- `internal/api/v1/` — HTTP handlers organized by domain: `files/`, `diary/`, `configuration/`
- `internal/app/` — Application bootstrap: `app.go` (init), `context.go` (dependency wiring), `routes.go` (route registration)
- `internal/config/` — Config loaded from `.env` via godotenv; build-tag-based path resolution
- `internal/worker/` — Background goroutine pool (200 workers) for async file processing tasks
- `pkg/database/` — DB connection, migrations, `DbContext` wrapper
- `pkg/logger/`, `pkg/i18n/`, `pkg/utils/` — Shared packages

**Build tags** determine path resolution at compile time:
- `dev` — paths relative to project root (for local development)
- `windows` — paths under `%ProgramFiles%\Kuranas\`
- `linux` — paths under `/etc/kuranas/`

**File processing pipeline** (worker): Directory walk → DTO conversion → Metadata extraction → Checksum computation → Database persistence. Each stage is a set of goroutines communicating over channels. Python scripts in `scripts/` handle audio/image/video metadata extraction.

**Backend env vars** (`.env` in `backend/`):
- `ENTRY_POINT` — root directory to scan for files
- `LANGUAGE` — locale (e.g. `pt-BR`)
- `ENABLE_WORKERS` — set to `"true"` to enable background file scanning
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — PostgreSQL connection

### Frontend (`frontend/`)

**Stack**: React 19, TypeScript, Vite, MUI (Material UI), TanStack Query, React Router v7

- `src/app/App.tsx` — Root routing (React Router `<Routes>`)
- `src/components/providers/appProviders.tsx` — Global provider tree: QueryClient, I18n, Snackbar, MUI dark theme, BrowserRouter
- `src/components/hooks/` — Feature-specific context providers (file, music, video, image, etc.) used as hook-based state managers
- `src/service/index.ts` — Axios instance configured with `VITE_API_URL` base URL (`/api/v1`)
- `src/pages/` — Page-level components (files, music, videos, images, analytics, etc.)

Path alias: `@/` → `src/`

**Frontend env var**: `VITE_API_URL` — points to the backend (e.g. `http://localhost:8000`). Configured in `.env`, `.env.development`, or `.env.production`.

TanStack Query is configured with `staleTime: 5 minutes` and no refetch on window focus/reconnect.

## Build, Test, and Development Commands

### Backend (run from `backend/`)

```bash
make -C backend run          # Run in dev mode (resolves paths relative to project root)
make -C backend test         # Run tests (-tags=dev)
make -C backend build        # Build (cross-compiles to Windows exe via mingw)
```

Backend tests use the `-tags=dev` build tag, which affects path resolution.

#### Run a single backend test file

```bash
cd backend && go test -tags=dev -v ./tests/files_test/...
cd backend && go test -tags=dev -v ./tests/diary_test/...
```

#### All backend tests with coverage

```bash
cd backend && go test ./... -cover
```

### Frontend (run from `frontend/`)

```bash
yarn dev           # Start Vite dev server (uses .env.development)
yarn test          # Run Jest tests
yarn test:watch    # Run Jest in watch mode
yarn lint          # Run ESLint
yarn build         # TypeScript check + Vite build
yarn coverage      # Run Jest with coverage report
```

### Mobile (run from `mobile/`)

```bash
./gradlew assembleDebug          # Build debug APK
./gradlew assembleRelease        # Build release APK
./gradlew test                   # Run unit tests
./gradlew connectedAndroidTest   # Run instrumented tests on connected device/emulator
```

### Full production build

```bash
make               # Builds frontend + backend, moves artifacts to build/
make clean         # Remove build artifacts
```

## Coding Style & Naming Conventions
Go code must be `gofmt`-clean and pass `go vet` (enforced in CI). Use lowercase package names, `CamelCase` exported identifiers, and keep feature code grouped by API domain under `internal/api/v1/<feature>`.

Frontend uses TypeScript + ESLint (`frontend/eslint.config.js`). Follow existing local style (including current tab-based indentation where present), use `PascalCase` for React components, and keep tests as `*.test.ts(x)`. Preserve path alias usage like `@/service/...`.

## Engineering Principles
All project code (backend and frontend) must follow Clean Code principles and programming best practices.

This is mandatory for every change in this repository:

- Follow Clean Architecture principles in addition to Clean Code.
- Keep pages, components, handlers, services, and functions small, cohesive, and focused on a single responsibility.
- Prefer clear separation of concerns, explicit boundaries, dependency inversion, and low coupling with high cohesion.
- Avoid god objects, oversized files, hidden side effects, duplicated logic, speculative abstractions, and hard-to-maintain control flow.
- Use names that reveal intent, keep functions at a single level of abstraction, and make behavior easy to read, test, and change.
- New code must be structured so it is easy to test, with tests covering the changed behavior and relevant edge cases.
- Every delivered change must include 100% test coverage for the code that was added or modified.
- No change may be considered complete if any existing test fails, even outside the changed area.
- No change may be considered complete if any quality gate fails, including tests, coverage thresholds, lint, formatting, type checks, or any other configured repository validation.
- Delivery is only acceptable when the repository is fully passing the applicable quality gates for the affected stack(s).

## Internationalization (Mandatory)
This project uses a single i18n source of truth based on JSON `key:value` translation files.

- Never hardcode user-facing text in backend or frontend code.
- Every user-visible string must be added to the i18n JSON files first.
- Backend must consume i18n strings directly from these JSON files.
- Frontend must fetch translations from backend and use the same JSON keys/content.
- Any new feature or change that introduces/updates visible text must include the corresponding i18n JSON updates.

## Testing Guidelines
Frontend tests run on Jest + Testing Library (`jsdom`) with global minimum coverage thresholds of 80% for branches/functions/lines/statements. Backend tests use Go's `testing` package; place tests as `*_test.go` alongside package code or under `backend/tests`.

## Commit & Pull Request Guidelines
Use Conventional Commit style seen in history, e.g. `feat(images): ...`, `fix(ui): ...`, `refactor(frontend): ...`. Keep subject lines imperative and scoped.

PRs should include:
- clear behavior summary and motivation;
- linked issue (if available);
- UI screenshots/GIFs for visual changes;
- commands run and test/lint results.

Ensure `.github/workflows/quality.yml` checks pass before requesting review.

## Frontend Standards (Persistent Reference)
For any frontend task, always consult `docs/standards/frontend-standards.md` before reading/changing implementation files.

If there is a conflict between existing code and this standards file, follow the standards file and keep changes consistent with adjacent code where safe.

## Backend Standards (Persistent Reference)
For any backend task, always consult `docs/standards/backend-standards.md` before reading/changing implementation files.

If there is a conflict between existing code and this standards file, follow the standards file and keep changes consistent with adjacent code where safe.

## Mobile Standards (Persistent Reference)
The mobile app targets **Android 4.1.2 (API 16)** on the **Samsung Galaxy Tab 2 7.0 (GT-P3110)**. The mandatory stack is **Java + XML Views + AppCompat**. Kotlin and Jetpack Compose are forbidden — Compose requires API 21+, and Kotlin adds unnecessary runtime overhead and potential Dalvik incompatibilities on API 16. All Android API calls and libraries must be compatible with API level 16. Validate device-specific constraints (screen size 1024×600, 7" form factor, hardware capabilities) when making layout or feature decisions.

## Task Management Protocol
Development planning docs live at `/home/server/Documentos/docs/kuranas/`.
Structure: `analysis/` (large research) → `decisions/` (refined, <100 lines) → `tasks/` (executable, <80 lines).

- `tasks/backlog.md` is the single source of truth for task ordering.
- Only ONE task can be in `tasks/active/` per project at a time.
- Always update `backlog.md` AND `index.md` when changing task status.
- Task files are self-contained: include all context needed to execute.
- Never modify `analysis/` or `decisions/` files during task execution.
- When completing a task, add completion date and summary to the Done section.
- Start work: `/task-next`. Finish work: `/task-done`.
- Do NOT read `analysis/` or `decisions/` unless explicitly creating new tasks.
