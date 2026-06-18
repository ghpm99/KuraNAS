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

**Always split work into logical commits** — one coherent change per commit (e.g. backend / frontend / android / docs separately), never one giant mixed commit. **Do not add a `Co-Authored-By` trailer** to commit messages.

## Root build & quality gates (root `Makefile`)

- `make ci` — every gate (frontend + backend). Run before committing.
- `make ci-backend` — `gofmt` check, `go vet`, `go test ./...` with **coverage ≥ 80%** enforced.
- `make ci-frontend` — `yarn lint`, `yarn test --coverage`, `yarn build`, `yarn typecheck:test`.
- `make all` — builds frontend, cross-compiles backend (Windows), assembles `build/` (frontend `dist/` + `kuranas` binary + `icons/` + `translations/`), then runs the local `deploy` target.
- `make release-main-ff` — fetches, fast-forwards `develop` from `origin/main`, then fast-forwards `main` from `develop`. Requires a clean working tree.

`make deploy` and `make all`'s deploy step include `Makefile.local` (gitignored, optional).

**Only declare a delivery or fix done after `make ci` runs fully green.** `make ci` is slow, so don't run it on every commit — but at the end of a delivery or fix, run it to consolidate, and only call the work complete once every gate passes.

## The HTTP contract is the integration point

Because the apps are otherwise decoupled, a change to a backend route or DTO shape can break the frontend, both Android apps, and the plugin at once. When changing anything under `backend/internal/api/v1/`, check the consumers: frontend `src/service/*.ts`, the Android `feature/*/data` layers, and `plugin/src/background/uploader.js` (captures endpoints).

**Keep endpoints small.** One endpoint owns one piece of information and returns the smallest meaningful payload, via handler → service → repository with small functions and one optimized `.sql` per query. Do not build fat aggregate responses — the `analytics` feature (one endpoint per concern: `/analytics/storage`, `/analytics/types`, `/analytics/duplicates`, …, composed client-side) is the reference shape; never reintroduce an aggregate "overview" endpoint. Full rule in `backend/CLAUDE.md` → "Endpoint granularity & response shape".

## Backend domains: a generic file core + type extensions

Backend domains live as sibling packages under `backend/internal/api/v1/<domain>/`, **organized by domain, never by layer** (no `handlers/`/`services/` packages — layer is a *filename prefix*, not a folder). The file/media domains use a **supertype → extension** shape that mirrors the DB (one `files` table + per-type complement tables):

- `files/` is the **generic core** — it owns `FileModel`/`FileDto` and only generic behavior (CRUD, tree, listing, operations, recent, reports, generic blob/thumbnail) and **must not import** `image/`/`music/`/`video/` or know type-specific concepts.
- `image/`, `music/`, `video/` are **extensions** — each owns its complement table + specialized logic and **imports `files`**. Dependency flows one way (`extension → files`); a cycle means the modeling is wrong.
- **A package does not own a table.** Extensions freely `JOIN` the `files` table. "Everything is a file in the DB" is a data fact, not a reason to pile type logic into `files`.
- Cross-type screens **compose at the edge** (frontend, or a handler that may import several domains), never by the core reaching into extensions.

Full rule + migration status in `backend/CLAUDE.md` → "Domain package organization" and `docs/refactor/`.

## No user-facing literal strings — i18n is mandatory

**Every string a user can see must come from the app's i18n layer, never a hard-coded literal.** This covers labels, buttons, titles, placeholders, empty/loading states, toasts, and error/warning messages alike. Add the term to the translation catalog and reference it by key — never type the visible text inline.

Each app resolves text its own way:

| App | Mechanism | Keys live in |
|---|---|---|
| `backend/` | `i18n.GetMessage(key)` / `i18n.Translate(key, args…)` | `backend/translations/{pt-BR,en-US}.json` |
| `frontend/` | `useI18n().t(key, {var})` (`{{var}}` interpolation) | the same backend JSON, fetched via `/translations` (`getTranslations`) |
| `android/` | `stringResource(R.string.key)` | `app/src/main/res/values*/strings.xml` |
| `mobile/` | `TranslationManager` by key | `app/src/main/assets/translations/*.json` |

**Backend-originated messages are translated once, at the source, and shown as-is.** The backend resolves the string server-side (single locale from the `LANGUAGE` env) and returns the final text in `{"error": "..."}`. Clients display that text directly — they do **not** re-translate it and must **not** invent a parallel client string for it. So every client screen is a *mix*: its own static text goes through the client i18n, while server messages are rendered verbatim. We deliberately accept that the app and server may briefly be in different languages (decided: "mix simples").

Anti-patterns to fix on sight: a literal in JSX/Compose/Go that reaches the screen; `gin.H{"error": err.Error()}` (leaks an untranslated raw Go error — wrap it in an i18n key instead).

Plugin (`plugin/`) is **not yet** under this rule — revisit if/when it grows real user-facing UI.
</content>
</invoke>
