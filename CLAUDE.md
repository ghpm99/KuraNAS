# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KuraNAS is a personal NAS (Network Attached Storage) system with a Go backend and React/TypeScript frontend. The backend serves the frontend as a SPA and exposes a REST API at `/api/v1`.

## Commands

### Backend (run from `backend/`)

```bash
# Run in dev mode (resolves paths relative to project root)
make -C backend run

# Run tests
make -C backend test

# Build (cross-compiles to Windows exe via mingw)
make -C backend build
```

Backend tests use the `-tags=dev` build tag, which affects path resolution.

### Frontend (run from `frontend/`)

```bash
yarn dev           # Start Vite dev server (uses .env.development)
yarn test          # Run Jest tests
yarn test:watch    # Run Jest in watch mode
yarn lint          # Run ESLint
yarn build         # TypeScript check + Vite build
```

### Full production build

```bash
make               # Builds frontend + backend, moves artifacts to build/
```

### Run a single backend test file

```bash
cd backend && go test -tags=dev -v ./tests/files_test/...
cd backend && go test -tags=dev -v ./tests/diary_test/...
```

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
