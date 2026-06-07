# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Go 1.24 + Gin HTTP server, module `nas-go/api`. The single source of truth for the whole product; also serves the built web UI from `./dist`.

## Commands (`cd backend`)

- `make run` — dev server: `go run -tags=dev cmd/nas/main.go`, listens on `:8000`.
- `make test` — `go test -tags=dev -v ./...`.
- `make coverage` — coverage profile + `coverage.html`.
- `make build` — production build, **cross-compiles to Windows** (`GOOS=windows`, `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc`), injects version ldflags, outputs `kuranas.exe`.
- Single test: `go test -tags=dev ./internal/api/v1/files/ -run TestName`.

The root `make ci-backend` enforces `gofmt`, `go vet`, and **coverage ≥ 80%**.

## Build tags select environment

The same code compiles for three targets via build tags; `cmd/nas/` and `internal/config/` each have one file per tag:
- `dev` — `main.go` + `build_config.go`; paths relative to project root. This is what `make run`/`make test` use.
- `windows && !dev` / `linux && !dev` — production. The Windows entrypoint (`main_windows.go`) runs as a `kardianos/service` (Windows service), not a bare `main`.

## Composition root — everything is wired by hand (no DI framework)

- `internal/app/context.go` — `NewContext(*sql.DB)` builds the entire object graph into one `AppContext`. Each feature has a `new<Feature>Context` helper constructing **`repository → service → handler`** in that order, storing all three (plus interfaces) on a `<Feature>Context` struct hanging off `AppContext`.
- `internal/app/app.go` — `InitializeApp()` loads config/translations, opens the DB, builds the context, registers routes, then starts: a pool of **200** workers (`StartWorkers`), the folder watcher (60s poll), a UDP discovery listener, and an mDNS registrar (all on port 8000). Returns an `Application` you `Run()`/`Stop()`.
- `internal/app/routes.go` — `RegisterRoutes` mounts every feature under `/api/v1`, plus CORS (origins from `ALLOWED_ORIGINS`, `AllowCredentials: true`), gzip, Swagger, and the SPA fallback: `/assets` served static with immutable cache, `NoRoute → ./dist/index.html`. **This is why the backend serves the frontend.** Route registrars are nil-guarded — a missing feature context skips its routes rather than panicking.

To add a feature: a package under `internal/api/v1/<feature>/` (`handler.go`, `service.go`, `repository.go`, `interfaces.go`, `model.go`, `dto.go`), a `new<Feature>Context` in `context.go`, and a `Register<Feature>Routes` in `routes.go`.

Existing feature modules: `files`, `diary`, `music`, `video`, `analytics`, `jobs`, `configuration`, `search`, `notifications`, `captures`, `libraries`, `watchfolders`, `takeout`, `aiproviders`, `ollama`, `updater`, `distribution`, `health`.

`distribution` is filesystem-backed (no DB): it serves pre-built client apps (Android APKs, the browser-extension zip) from a `./downloads/` directory described by `downloads/manifest.json` — `GET /api/v1/downloads` lists them, `GET /api/v1/downloads/:id` streams one. Artifacts are built by `scripts/build-downloads.sh` (CI/maintainer, never the server), bundled into `build/` by the root `make move`, and synced in place by the `updater`. A missing `downloads/` directory simply yields an empty catalog.

## Layering & testing seams

Each feature defines `RepositoryInterface`, `ServiceInterface`, and any collaborator interfaces in `interfaces.go`. Handlers depend on the service interface; services on the repository interface — that is the mocking boundary. Repository tests use `github.com/DATA-DOG/go-sqlmock`. `*Model` is the DB shape, `*Dto` is the API/transport shape, and **conversion happens in the service layer**. Repositories are the only layer touching the DB; handlers only parse/validate requests and shape responses.

## Endpoint granularity & response shape (mandatory rule)

Every endpoint follows **handler → service → repository**, built from **small, single-purpose functions**, and **returns the smallest meaningful payload** — one endpoint owns one piece of information.

- **One resource per endpoint.** Do not aggregate unrelated data into a single "overview"/"dashboard" response. Each distinct concern (storage KPIs, type distribution, recent files, duplicates, processing queue, …) is its own route, its own handler, its own service method, its own repository method, and its own `.sql` file.
- **One small, optimized query per repository call.** No god-query and no fan-out of many queries behind a single endpoint. Each `.sql` file under `pkg/database/queries/<domain>/` answers exactly one question and stays small.
- **Why:** a fat response makes the app feel slow (the client waits for everything to render anything), couples consumers to fields they don't use, and means one broken detail takes the whole endpoint down and is hard to isolate. Small endpoints fail in isolation, are individually cacheable, and let the frontend load progressively.
- **Reference implementation:** the `analytics` feature is the canonical example. It used to expose one fat `GET /analytics/overview` returning a giant `OverviewDto` from ~12 SQL calls; it is now split into one endpoint per concern — `/analytics/storage`, `/analytics/timeseries`, `/analytics/types`, `/analytics/extensions`, `/analytics/recent-files`, `/analytics/top-folders`, `/analytics/hot-folders`, `/analytics/duplicates`, `/analytics/duplicates/groups`, `/analytics/library`, `/analytics/processing`, `/analytics/health`, `/analytics/insights` — each with its own handler/service/repository method and a single `.sql`. Follow this shape; never reintroduce an aggregate "overview" endpoint. The frontend composes these slices in `analyticsProvider` via independent queries, so each one loads/fails on its own.

## User-facing text goes through i18n (mandatory)

Any string that can reach a user — API `error`/message fields, notification titles/bodies, user-facing logged events — must come from `pkg/i18n`: `i18n.GetMessage("KEY")` (static) or `i18n.Translate("KEY", args…)` (with `%s`/`%d` placeholders), with the term added to **both** `translations/pt-BR.json` and `translations/en-US.json`. The active locale is the `LANGUAGE` env, loaded once at boot. **Never return `gin.H{"error": err.Error()}`** — that leaks a raw, untranslated Go error to the client; map it to an i18n key instead. Clients receive the already-translated text and render it verbatim. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".

## Database

- PostgreSQL via `lib/pq`; connection from `DB_*` env vars (`pkg/database`).
- `pkg/database/dbContext.go` — `DbContext` wraps `*sql.DB` with `ExecTx`/`QueryTx` transaction helpers; repositories receive a `*DbContext`.
- **SQL is never inline.** Each query is a `.sql` file under `pkg/database/queries/<domain>/`, embedded into a sibling `<domain>.go` via `//go:embed` into an exported `var`. Add a query by dropping a `.sql` file + a `//go:embed` line. Domains today: `aiproviders`, `analytics`, `captures`, `configuration`, `diary`, `file`, `jobs`, `libraries`, `log`, `music`, `notifications`, `search`, `systemevent`, `video`, `watch_folders`.
- Migrations: numbered `.sql` files in `pkg/database/migrations/queries/`, applied on startup. Schema changes go here, never as ad-hoc DDL.

## Worker subsystem (`internal/worker`)

Two coexisting execution models, both started by `StartWorkers` (gated by `ENABLE_WORKERS`):
1. **Legacy task channel** — `chan utils.Task` consumed by the worker pool.
2. **Job/Step orchestrator (preferred)** — persists a DAG of jobs and steps to `worker_job*` tables. Enumerated in `job_domain.go`, each with an `IsValid()` guard:
   - Job types: `startup_scan`, `upload_process`, `fs_event`, `reindex_folder`, `takeout_import`, `ollama_pull`.
   - Step types: `scan_filesystem`, `diff_against_db`, `metadata`, `checksum`, `persist`, `thumbnail`, `playlist_index`, `mark_deleted`, `takeout_extract`, `ollama_model_pull`.
   - Steps carry `DependsOn` and `MaxAttempts`; jobs/steps have priority (`low`/`normal`/`high`, weighted) and status enums.

## AI subsystem (`pkg/ai`) — hot-swappable

- Providers under `pkg/ai/providers/{ollama,openai,anthropic}` implement `ai.Provider`. An `ai.Router` maps each `TaskType` (classification, extraction, summarization, generation, simple, complex) to a priority-ordered provider chain with retry/fallback (`WithRetry`).
- The app holds an `ai.Manager` (the `ai.ServiceInterface`) that can `Swap` the active service at runtime.
- `context.go::newAIStack` rebuilds the chain from the `ai_providers` DB table whenever config changes (registered via `service.SetOnChange`). **Operational tuning (model, base_url, timeout, retries) lives in the DB; only API keys come from env: `AI_OPENAI_API_KEY`, `AI_ANTHROPIC_API_KEY`.** A cloud provider with no key is logged and skipped. The Ollama daemon base URL is resolved dynamically from the DB config too.
- Prompts are `.txt` files under `pkg/ai/prompts/`, embedded via Go (`prompts.go`).

## Other cross-cutting pieces

- `internal/discovery` — mDNS registrar + UDP listener advertising the server on port 8000 for LAN auto-discovery.
- `internal/watcher` — polling folder watcher (60s) driving auto-organization (libraries, watch folders, Google Takeout import).
- `pkg/`: `i18n` (translation files), `logger`/`systemevent` (DB-backed), `icons`/`img`/`pdf` (thumbnails & media), `utils` (generic `PaginationResponse[T]`, tasks).
- Config is env-driven (`internal/config`, loaded from `.env` via `godotenv`). Env vars: `ENTRY_POINT` (root dir KuraNAS indexes), `ENABLE_WORKERS`, `LANGUAGE`, `ENV`, `ALLOWED_ORIGINS`, `DB_HOST/PORT/USER/PASSWORD/NAME`. `ToRelativePath`/`ToAbsolutePath` translate between stored relative paths and disk paths.
</content>
