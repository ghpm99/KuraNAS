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

## API base URL resolution (`src/service/apiUrl.ts`)

The same build talks to the bundled backend (same origin) or a remote dev server. `getApiBaseUrl()` resolves in order:
1. runtime global `__KURANAS_API_URL__`,
2. `VITE_API_URL` (via `@/config/viteEnv`),
3. `process.env.VITE_API_URL` (for Jest/Node),
4. empty → `getApiV1BaseUrl()` falls back to same-origin `/api/v1`.

All service modules use the shared axios instance `apiBase` (`src/service/index.ts`), built on that base URL. There is **one service module per API domain** under `src/service/` (`files`, `music`, `playlist`, `videoPlayback`, `analytics`, `jobs`, `notifications`, `libraries`, `takeout`, `aiProviders`, `ollama`, `configuration`, `search`, `activityDiary`, `playerState`, `update`, `downloads`). These mirror the backend routes — keep them in sync when backend DTOs change.

## App structure

- `src/app/App.tsx` — route table, **all pages lazy-loaded**, wrapped `AppProviders → ErrorBoundary → GlobalMusicProvider`, with a persistent `GlobalPlayerControl` mini-player (hidden on the video-player route). Route paths live in `src/app/routes`.
- Heavier domains live under `src/features/{files,music,videos}/` and own **all** their domain UI (providers, views, components — e.g. `features/music/providers/GlobalMusicProvider`, `features/music/components/player/GlobalPlayerControl`). A domain's code never lives under `src/components/`.
- Other dirs: `src/pages/<page>/` (route shells), `src/components/` (**shared/cross-domain UI only** — layout, tabs, search, settings…), `src/service/`, `src/types/`, `src/theme/`, `src/shared/`, `src/utils/`, `src/config/`.
- Path alias `@` → `src` (configured in both `vite.config.ts` and `jest.config.js`'s `moduleNameMapper`).

## User-facing text goes through i18n (mandatory)

No hard-coded literal may reach the screen — labels, buttons, placeholders, empty/loading states, toasts, error/warning messages. Use `const { t } = useI18n()` and `t("KEY", { var })` (`{{var}}` interpolation); add the term to the backend translation JSON (`backend/translations/*.json`, served to the app via `/translations`). **Exception:** messages returned by the backend (e.g. an axios `error.response.data.error`) arrive **already translated** — render them as-is; don't wrap them in `t()` or duplicate the string as a frontend key. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".

## Stack notes

- MUI v7 (`@mui/material`, `x-charts`, `x-data-grid`), TanStack Query, `react-router-dom` v7, axios, framer-motion, notistack, lucide-react.
- `vite.config.ts` uses `@vitejs/plugin-legacy` (targets include `Opera >= 50`, modern polyfills on) and manually chunks vendors (`vendor-mui-x`, `vendor-mui`, `vendor-motion`, `vendor-query`).
- Tests: Jest + ts-jest + Testing Library (jsdom); `*.test.tsx?` colocated with code. `@/config/viteEnv` is mapped to `viteEnv.jest.ts` under test so `import.meta.env` doesn't break in Node.
</content>
