# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

React 19 + Vite 7 + TypeScript web UI (Yarn 1, `packageManager` pinned to `yarn@1.22.22`). Its production build is bundled into the backend binary's working dir and served same-origin.

## Commands (`cd frontend`)

- `yarn dev` — Vite dev server.
- `yarn build` — `tsc -b && vite build`.
- `yarn lint` — ESLint over the repo.
- `yarn test` — Jest, runs `--runInBand`. `yarn coverage` adds coverage.
- `yarn typecheck:test` — type-checks test files against `tsconfig.test.json` (`tsc -p tsconfig.test.json --noEmit`).
- Single test: `yarn test src/service/files.test.ts` or `yarn jest -t "test name"`.

Coverage thresholds are enforced in `jest.config.js`: **branches 89, functions 90, lines 90, statements 90**. `src/service/index.ts` is excluded from coverage.

## Component tests are mandatory — resilience-first (owner rule)

**Every component must have a test, and its first test is a plain render with NO service/backend mock.** The component must render without throwing when the backend is absent or returns a partial/empty payload — a component may never assume the server sent a complete, well-shaped response. This is the guard against the whole-tree crash we hit when a settings group came back `undefined` (`settings.players` / `settings.appearance`).

Concretely:
- The first `it(...)` renders the component (or hook) with services left unmocked / rejecting, and asserts it mounts and shows something sane (loading/empty/fallback) instead of crashing.
- Make the component tolerant at the data edge: default props, optional chaining, and **fill missing groups from a default** before use. The reference fix is `mergeSettingsConfiguration` in `components/providers/settingsProvider/index.tsx` (backfills every group from `defaultSettingsConfiguration`); the regression test is the "partial payload" case in its `index.test.tsx`.
- Only after the no-mock render passes do the behavior tests (with mocked services) follow.

Coverage thresholds (branches 89 / functions 90 / lines 90 / statements 90, in `jest.config.js`) still apply on top of this.

## API base URL resolution (`src/service/apiUrl.ts`)

The same build talks to the bundled backend (same origin) or a remote dev server. `getApiBaseUrl()` resolves in order:
1. runtime global `__KURANAS_API_URL__`,
2. `VITE_API_URL` (via `@/config/viteEnv`),
3. `process.env.VITE_API_URL` (for Jest/Node),
4. empty → `getApiV1BaseUrl()` falls back to same-origin `/api/v1`.

All service modules use the shared axios instance `apiBase` (`src/service/index.ts`), built on that base URL. There is **one service module per API domain** under `src/service/` (`files`, `music`, `playlist`, `videoPlayback`, `analytics`, `jobs`, `notifications`, `libraries`, `takeout`, `aiProviders`, `ollama`, `configuration`, `search`, `activityDiary`, `playerState`, `update`, `downloads`). These mirror the backend routes — keep them in sync when backend DTOs change.

## App structure

- `src/app/App.tsx` — route table, **all pages lazy-loaded**, wrapped `AppProviders → ErrorBoundary → GlobalMusicProvider`, with a persistent `GlobalPlayerControl` mini-player (hidden on the video-player route). Route paths live in `src/app/routes`. **The `AppShell` is mounted once, by a layout route (`<Route element={<ShellLayout/>}>` with `<Outlet/>`)** — pages and feature layouts never wrap themselves in the shell; the video player is the only route outside it.
- Heavier domains live under `src/features/{files,music,videos}/` and own **all** their domain UI (providers, views, components — e.g. `features/music/providers/GlobalMusicProvider`, `features/music/components/player/GlobalPlayerControl`). A domain's code never lives under `src/components/`.
- Other dirs: `src/pages/<page>/` (route shells), `src/components/` (**shared/cross-domain UI only** — layout, tabs, search, settings…), `src/service/`, `src/types/`, `src/theme/`, `src/shared/`, `src/utils/`, `src/config/`.
- Path alias `@` → `src` (configured in both `vite.config.ts` and `jest.config.js`'s `moduleNameMapper`).

## Page layout system (issue #112 — no hero cards)

Every screen is built from the shared layout primitives in `src/components/layout/`; sizes come from tokens in `src/theme/visualTokens.ts` (`--app-font-page-title`, `--app-content-max-width`, spacing/radius vars) — never hard-code a page-title font size or page padding:

- **`PageContainer`** — the page wrapper: max-width, padding, vertical gap. Screens do not define their own `.page` padding.
- **`PageHeader`** — the only page header: compact `h1` + optional one-line subtitle + `actions` slot. **Never reintroduce a "hero" card** (eyebrow + giant title + description in a gradient box) — that pattern was removed on purpose; a screen's description is a short subtitle, not a panel.
- **`DomainPageLayout`** (`header` + `nav` + children) with **`DomainNavTabs`** — domain sections (Images/Music/Videos/Analytics) navigate via a tab strip under the header, **not** a second sidebar. The whole page scrolls in the shell's scroll area: no `overflow: hidden` page grids or nested scroll panes.
- One `h1` per page (the `PageHeader`); inner section titles are `h2`+.

## User-facing text goes through i18n (mandatory)

No hard-coded literal may reach the screen — labels, buttons, placeholders, empty/loading states, toasts, error/warning messages. Use `const { t } = useI18n()` and `t("KEY", { var })` (`{{var}}` interpolation); add the term to the backend translation JSON (`backend/translations/*.json`, served to the app via `/translations`). **Exception:** messages returned by the backend (e.g. an axios `error.response.data.error`) arrive **already translated** — render them as-is; don't wrap them in `t()` or duplicate the string as a frontend key. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".

## Stack notes

- MUI v7 (`@mui/material`, `x-charts`, `x-data-grid`), TanStack Query, `react-router-dom` v7, axios, framer-motion, notistack, lucide-react.
- `vite.config.ts` uses `@vitejs/plugin-legacy` (targets include `Opera >= 50`, modern polyfills on) and manually chunks vendors (`vendor-mui-x`, `vendor-mui`, `vendor-motion`, `vendor-query`).
- Tests: Jest + ts-jest + Testing Library (jsdom); `*.test.tsx?` colocated with code. `@/config/viteEnv` is mapped to `viteEnv.jest.ts` under test so `import.meta.env` doesn't break in Node.
</content>
