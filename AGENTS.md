# Repository Guidelines

## Project Structure & Module Organization
`backend/` contains the Go API and worker pipeline. Main entrypoint is `backend/cmd/nas/main.go`; HTTP modules live under `backend/internal/api/v1`; shared infrastructure is in `backend/pkg` (database, i18n, utils, media helpers). SQL migrations and query files are in `backend/pkg/database/migrations` and `backend/pkg/database/queries`.

`frontend/` is a Vite + React + TypeScript app. Core UI code is in `frontend/src/components`, route-level pages in `frontend/src/pages`, API clients in `frontend/src/service`, and shared types/utilities in `frontend/src/types` and `frontend/src/utils`.

`build/`, `frontend/dist/`, and `frontend/coverage/` are generated artifacts.

## Build, Test, and Development Commands
- `make`: builds frontend and backend, then assembles distributable files under `build/`.
- `make clean`: removes build artifacts.
- `make -C backend run`: runs backend in dev mode with build metadata flags.
- `cd backend && go test ./... -cover`: runs all backend tests with coverage.
- `make -C backend test`: runs file-scanning tests (`-tags=dev`).
- `cd frontend && yarn dev`: starts Vite dev server.
- `cd frontend && yarn build`: type-checks and builds production assets.
- `cd frontend && yarn lint`: runs ESLint.
- `cd frontend && yarn test --watchAll=false` or `yarn coverage`: runs Jest tests/coverage.

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
Frontend tests run on Jest + Testing Library (`jsdom`) with global minimum coverage thresholds of 80% for branches/functions/lines/statements. Backend tests use Go’s `testing` package; place tests as `*_test.go` alongside package code or under `backend/tests`.

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
