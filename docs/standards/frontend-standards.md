# Frontend Standards (Canonical)

This document is the canonical source for frontend implementation patterns in this repository.

## Agent Enforcement
- Before any frontend change, read this file first.
- If a rule here conflicts with local file style, prefer this file and then normalize surrounding code when safe.
- If a requested change conflicts with these rules, explicitly call out the conflict before implementing.

## 1) Architecture
- Stack is `React + TypeScript + Vite`, with route composition in `src/app/App.tsx`.
- Keep route-level files in `src/pages/*` and keep page wrappers thin.
- For `files`, `music`, and `videos`, business behavior must live in `src/features/<domain>/*` (providers/hooks/views/components).
- Keep `src/components/*` focused on shared UI composition, layout, and domains not yet migrated to feature-first ownership.
- Keep feature boundaries explicit: each major domain (`files`, `music`, `videos`, `analytics`, `activityDiary`, `about`, `images`) owns its components, provider/hooks, and tests.
- Use path alias imports (`@/...`) for internal modules instead of deep relative traversals.
- All system logic and data-fetching logic must be implemented in hooks/providers, not in render components.
- Apply single responsibility strictly: pages and presentational components must stay small and focused.

## 2) Components
- Use `PascalCase` for React component names and component files when the file exports a single component (`Button.tsx`, `Layout.tsx`).
- Prefer thin wrapper pages/components (`index.tsx`) that compose providers + screen components, keeping rendering flow easy to read.
- Keep shared UI primitives in `src/components/ui/*` and domain components under their feature folder.
- For conditional rendering, prefer early returns and simple guard branches.
- User-facing text must come from `t('KEY')`; never inline literal UI strings.
- Component implementation pattern must follow this abstraction whenever applicable:
- `ComponentName.module.css` (styles), `useComponentName.ts` (logic), `ComponentName.tsx` (view composition).
- `ComponentName.tsx` should import the hook + CSS module and focus on wiring JSX only.
- Prefer existing components from project or MUI before creating new UI primitives.
- Do not recreate controls already available in MUI or in `src/components/ui/*`.

## 3) State Management
- Use React Query for server state (`useQuery`, `useInfiniteQuery`, `useMutation`) and React Context for UI/domain local state.
- Each context must expose a typed hook and throw when used outside provider (`useX must be used within XProvider` pattern).
- Memoize provider value objects and derived collections when they depend on frequently changing state.
- Use `queryKey` naming that is feature-oriented and stable (e.g. `['video-playlists']`, `['files', queryParams, filter]`).
- Any non-trivial business rule must live in hooks/providers, even if triggered by a small button or local UI event.
- For nested UI areas (example: toolbar with action buttons), centralize button behavior in that feature provider and consume via its hook.
- Simple UI-only calculations (e.g. icon visibility, small derived flags) may live in isolated `use*` hooks without creating a provider.

## 4) API/Service Layer
- All HTTP calls must go through `apiBase` (`src/service/index.ts`) built on axios.
- `apiBase` base URL must come from `getApiV1BaseUrl()` (`src/service/apiUrl.ts`); do not hardcode API roots.
- Keep API DTO types and request payload types colocated with each service module.
- Service functions should return `response.data` and keep transport mapping localized in the service layer.
- HTTP calls must execute inside providers/hooks only.
- Render components/pages must never perform HTTP calls directly.
- Components/providers should call service functions, not raw axios endpoints.

## 5) Styling/CSS
- Styling standard for new code:
- Use CSS Modules (`*.module.css`) as default for component and feature styling.
- Use MUI theme and component props for visual consistency/integration with library components.
- Avoid `sx` as primary styling strategy for new feature code.
- Do not create new plain CSS (`*.css`) files; keep plain CSS only where legacy code already exists and is not being refactored in that task.
- Keep spacing, borders, and color choices aligned with the dark theme defined in `src/components/providers/appProviders.tsx`.
- Preserve existing local formatting conventions (tabs where already used).

## 6) i18n (Mandatory)
- Never hardcode user-visible text in frontend components.
- All user-visible strings must come from backend-provided translation keys/content.
- Any feature introducing visible text must include i18n updates.
- Frontend translation access must go through `useI18n()` from `src/components/i18n/provider/i18nContext.ts`.
- Translation source is backend endpoint `/configuration/translation`; keys should be stable and semantic (e.g. `VIDEO_ADD_SUCCESS`).
- When interpolation is needed, use placeholder replacement via `t('KEY', { name: value })`.

## 7) Testing
- Test stack is `Jest + ts-jest + Testing Library` with `jsdom` environment.
- Keep tests colocated using `*.test.ts` and `*.test.tsx`.
- Maintain global coverage threshold at minimum 80% for branches/functions/lines/statements (`jest.config.js`).
- For entry/provider-heavy modules, mock external boundaries (router, root rendering, network) and assert behavior/output.
- Always run, at minimum, `yarn lint`, `yarn test --watchAll=false`; use `yarn coverage` when touching critical flows.

## 8) Accessibility
- Prefer semantic elements (`header`, `nav`, button types, list semantics) before adding ARIA-only fixes.
- Ensure actionable icons/controls expose accessible names (`title`, visible label, or translated text).
- Provide translated `alt` text for media where applicable.
- Preserve keyboard usability for navigation and dialogs/drawers (focusable controls, close handlers).

## 9) Performance
- Favor React Query cache/invalidation over ad-hoc refetch loops.
- Use global query defaults from `AppProviders` and override per-query only when freshness requirements differ.
- Memoize expensive derived lists/maps in providers (`useMemo`) and stable handlers (`useCallback`) when passed down.
- Use paginated/infinite loading patterns for large datasets (`page_size`, `useInfiniteQuery`, intersection observer).
- Avoid unnecessary rerenders by keeping provider context values stable and minimal.

## 10) Pull Request Checklist (Frontend)
- User-facing text added/changed: i18n keys/content updated and consumed via `t(...)`.
- API changes: service module + DTO/request typings updated in `src/service/*`.
- UI changes: responsive behavior validated (desktop + mobile breakpoints already used in layout/header/sidebar).
- Tests added/updated for behavior regressions; coverage remains above project threshold.
- Lint/test/build commands executed successfully for frontend scope.
- If a new pattern is introduced, update this standards file in the same PR.

## Change Log
- 2026-03-05: Initial canonical frontend standards file created.
- 2026-03-05: First populated version based on current frontend implementation.
- 2026-03-05: Rules tightened: logic/HTTP only in hooks/providers, component abstraction (`.module.css` + `use*` + view), CSS Modules as default, and mandatory reuse of existing/MUI components.
- 2026-04-01: Updated architecture guidance to reflect feature-first ownership for `files`, `music`, and `videos`.
