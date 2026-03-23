## 1. Feature Overview

* Name: Home Dashboard Aggregation
* Summary: Home screen that composes recent files, favorites, recent images, analytics snapshot, and music/video resume cards.
* Purpose: Provide a single “return point” showing what matters now.
* Business value: Improves daily engagement and reduces navigation steps across modules.

## 2. Current Implementation

* How it works today: Frontend-only orchestration hook (`useHomeScreen`) fan-outs multiple API queries and derives unified card models.
* Main flows:
  * Home hook fetches analytics, favorites, images, video home catalog/playback state, music player state/now-playing tracks.
  * Derived selectors compute resume progress percentages and card data.
  * Home page renders sections with loading states.
* Entry points (routes, handlers, jobs):
  * Frontend route `/home`.
  * Backend dependencies consumed: analytics, files, video playback/catalog, music player/playlist endpoints.
* Key files involved (list with paths):
  * `frontend/src/pages/home/index.tsx`
  * `frontend/src/components/home/useHomeScreen.ts`
  * `frontend/src/service/analytics.ts`
  * `frontend/src/service/files.ts`
  * `frontend/src/service/videoPlayback.ts`
  * `frontend/src/service/playerState.ts`
  * `frontend/src/service/playlist.ts`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend composition layer over multiple backend domains.
  * No dedicated backend home endpoint.
* Data flow (step-by-step):
  * Home route mounts and triggers parallel React Query requests.
  * Results are normalized/sliced to display limits.
  * Resume cards compute progress from current time/duration and fallback persisted state.
  * UI renders consolidated dashboard cards.
* External integrations:
  * None direct; depends on internal API endpoints.
* State management (if applicable):
  * React Query + GlobalMusic context.

## 4. Data Model

* Entities involved:
  * Aggregated home view model (recent files, favorites, images, video resume, music resume, analytics snapshot).
* Database tables / schemas:
  * Indirectly depends on underlying module tables (`home_file`, player states, analytics sources).
* Relationships:
  * Home model is a derived read model combining multiple domain DTOs.
* Important fields:
  * Progress percent and resume position/duration for media cards.

## 5. Business Rules

* Explicit rules implemented in code:
  * Home limits: favorites/images/recent slices and video-home limits.
  * Progress values are clamped to 0-100.
  * Music resume falls back from in-memory active track to persisted player state.
* Edge cases handled:
  * Missing state in any module degrades to null card rather than crash.
* Validation logic:
  * Derived model utilities guard against non-finite values.

## 6. User Flows

* Normal flow:
  * User opens Home -> sees personalized resume and recent activity cards -> deep-links into files/music/videos/analytics.
* Error flow:
  * Individual widget failures do not block all widgets; partial dashboard still renders.
* Edge cases:
  * Stale cross-module state can show inconsistent resume vs actual playback.

## 7. API / Interfaces

* Endpoints:
  * No dedicated home endpoint; consumes existing domain endpoints.
* Input/output:
  * Aggregated frontend model composed from multiple service response contracts.
* Contracts:
  * Home screen depends on backward compatibility of multiple APIs simultaneously.
* Internal interfaces:
  * `useHomeScreen` exported data contract to home page components.

## 8. Problems & Limitations

* Technical debt:
  * Fan-out query pattern increases coupling to many module contracts.
* Bugs or inconsistencies:
  * Partial stale data between queries can create temporary contradictory states.
* Performance issues:
  * Multiple network calls on initial home load.
* Missing validations:
  * No centralized schema validation at aggregation boundary.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code detected.
* External code execution:
  * None.
* Unsafe patterns:
  * Dashboard indirectly surfaces potentially sensitive operational metadata without access-control differentiation.
* Injection risks:
  * Aggregates data from many domains; all text fields should be treated as untrusted render content.
* Hardcoded secrets:
  * None.
* Unsafe file/system access:
  * None direct in frontend aggregation logic.

## 10. Improvement Opportunities

* Refactors:
  * Introduce backend-for-frontend home endpoint to reduce client fan-out and contract fragility.
* Architecture improvements:
  * Add typed aggregation schema with versioning.
* Scalability:
  * Server-side aggregation + caching could lower frontend latency.
* UX improvements:
  * Add personalization controls for visible widgets and ordering.

## 11. Acceptance Criteria

* Functional:
  * Home dashboard displays recent/favorite/media-resume/analytics cards with valid deep links.
* Technical:
  * Dashboard renders partial success when some data sources fail.
  * Progress calculations are bounded and deterministic.
* Edge cases:
  * Missing player/catalog state produces null-safe UI without crashes.

## 12. Open Questions

* Unknown behaviors:
  * Whether home composition should eventually move server-side.
* Missing clarity in code:
  * No explicit product contract for required vs optional widgets.
* Assumptions made:
  * Home is optimized for convenience, not strong consistency across all modules.
