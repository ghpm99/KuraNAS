## 1. Feature Overview
- Name: Home Dashboard
- Summary: Central hub that aggregates system health and quick access to files, favorites, images, music resume, and video resume.
- Target user: Primary NAS user managing personal library daily.
- Business value: Reduces navigation friction and increases daily engagement by surfacing high-value actions and “continue” contexts.

## 2. Problem Statement
- Users need one place to understand system state and resume media quickly.
- Without it, users must navigate multiple sections to check storage, recent content, and playback continuity.

## 3. Current Behavior
- Loads analytics snapshot (30d), favorites subset, recent images subset, video home catalog, video playback state, and music state/queue.
- Shows quick actions to major domains.
- Provides direct open behavior for files/media and opens global search from the hero input.
- Builds resume cards from persisted player state when active in-memory player state is unavailable.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/home`.
2. System fetches aggregated data from analytics/files/images/music/video APIs.
3. Dashboard renders cards (storage, indexed files, status), quick actions, and resume widgets.
4. User clicks an item (file/image/music/video) and system routes to player/domain context.

### 4.2 Alternative Flows
1. User opens quick action; dashboard routes directly to target page.
2. User opens global search from home and jumps to any domain entity.
3. If music queue is empty in-memory, dashboard uses backend player state + now playing playlist to reconstruct resume.

### 4.3 Error Scenarios
1. Any data source fails: corresponding section remains empty/loading while page still renders.
2. Video playback state missing: no video resume card is shown.
3. Music player state missing: no music resume card is shown.

## 5. Functional Requirements
- The system must aggregate multiple domain APIs into one home view.
- The user can launch global search from home.
- The system must expose direct navigation actions to core domains.
- The system should show media continuation context when playback state exists.

## 6. Business Rules
- Home analytics period is fixed to `30d`.
- Favorites and recent images are capped for dashboard display.
- Video continue block is sourced from playback state and catalog “continue” section.
- Music resume prefers active global player state, then persisted player state.

## 7. Data Model (Business View)
- Analytics Overview: storage/health/processing KPIs.
- File Item: id, name, path, format, size, timestamps.
- Image Item: file + image metadata/classification.
- Music State: current track, position, queue size.
- Video Session: playlist, current video, progress.

## 8. Interfaces
- User interfaces: `HomeScreen` (`/home`).
- APIs:
  - `GET /analytics/overview`
  - `GET /files/tree?category=starred`
  - `GET /files/images`
  - `GET /video/catalog/home`
  - `GET /video/playback/state`
  - `GET /music/player-state/`
  - `GET /music/playlists/now-playing`
  - `GET /music/playlists/:id/tracks`

## 9. Dependencies
- Analytics, files, images, music, video, global search, i18n.
- Settings (player persistence flags affect resume behavior).

## 10. Limitations / Gaps
- Partial failures are silent at card level (limited user guidance).
- Dashboard actions for analytics tables navigate generically to files, not deep filtered views.

## 11. Opportunities
- Add explicit card-level error/retry affordances.
- Add deep-link actions from analytics cards (duplicates, top files, recent).
- Add personalization for home card order.

## 12. Acceptance Criteria
- Given home is opened, when APIs return successfully, then dashboard shows at least one section per domain.
- Given video playback exists, when home loads, then a continue/resume element is displayed.
- Given no playback state, when home loads, then resume sections are hidden without blocking page render.
- Given user clicks quick action, when route resolves, then user lands on selected domain page.

## 13. Assumptions
- Home is intended as read/launch surface, not a full management surface.
- Card-level loading and partial degradation are intentional UX choices.
