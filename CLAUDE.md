# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Monorepo layout

KuraNAS is a self-hosted NAS. It is a monorepo of five independent applications that all talk to one HTTP API (`/api/v1`, served on port `8000`):

| Dir | Stack | Talks to API as |
|---|---|---|
| `backend/` | Go 1.24 + Gin (module `nas-go/api`) | The API itself; also serves the built web UI |
| `frontend/` | React 19 + Vite + TypeScript | Same-origin (bundled into backend) or remote dev server |
| `android/` | Kotlin + Jetpack Compose (`com.kuranas.android`) | Server chosen at runtime (LAN discovery) |
| `mobile/` | Android in Java (`com.kuranas.mobile`), `minSdk 16` | Compile-time `API_BASE_URL` |
| `plugin/` | MV3 browser extension ("KuraNAS Stream Grabber") | Uploads captures to `/captures/upload/*` |

Each app has its own build system and its own `CLAUDE.md` with stack-specific detail — read the one for the directory you're working in (`backend/CLAUDE.md`, `frontend/CLAUDE.md`, `android/CLAUDE.md`, `plugin/CLAUDE.md`, `mobile/CLAUDE.md`). The only thing tying the apps together is the root `Makefile` (backend + frontend) and the shared HTTP contract.

## Git workflow

**All changes are committed directly to the `develop` branch.** Do not create extra branches unless explicitly asked. `develop` is the integration branch; `make release-main-ff` fast-forwards `main` from it.

## Root build & quality gates (root `Makefile`)

- `make ci` — every gate (frontend + backend). Run before committing.
- `make ci-backend` — `gofmt` check, `go vet`, `go test ./...` with **coverage ≥ 80%** enforced.
- `make ci-frontend` — `yarn lint`, `yarn test --coverage`, `yarn build`, `yarn typecheck:test`.
- `make all` — builds frontend, cross-compiles backend (Windows), assembles `build/` (frontend `dist/` + `kuranas` binary + `icons/` + `translations/`), then runs the local `deploy` target.
- `make release-main-ff` — fetches, fast-forwards `develop` from `origin/main`, then fast-forwards `main` from `develop`. Requires a clean working tree.

`make deploy` and `make all`'s deploy step include `Makefile.local` (gitignored, optional).

## The HTTP contract is the integration point

Because the apps are otherwise decoupled, a change to a backend route or DTO shape can break the frontend, both Android apps, and the plugin at once. When changing anything under `backend/internal/api/v1/`, check the consumers: frontend `src/service/*.ts`, the Android `feature/*/data` layers, and `plugin/src/background/uploader.js` (captures endpoints).
</content>
</invoke>
