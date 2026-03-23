## 1. Feature Overview

* Name: Music Playback and Playlists
* Summary: Delivers music library browsing, manual/system playlists, now-playing queue, and persisted player state.
* Purpose: Provide a full music listening workflow inside the NAS experience.
* Business value: Increases media consumption time and turns stored audio into an active product surface.

## 2. Current Implementation

* How it works today: Backend manages playlist and catalog data; frontend uses global music provider and player-state sync to drive playback UI.
* Main flows:
  * User browses music catalog by home/artists/albums/genres/folders.
  * User creates and edits playlists; system playlists are generated/maintained.
  * Player state is fetched/saved by client IP (`client_id`).
  * Global music provider hydrates queue and syncs local playback with backend state.
* Entry points (routes, handlers, jobs):
  * Playlists: `GET/POST /api/v1/music/playlists`, `GET/PUT/DELETE /api/v1/music/playlists/:id`, `GET/POST/DELETE/PUT` track subroutes
  * System/now-playing: `GET /api/v1/music/playlists/system`, `GET /api/v1/music/playlists/now-playing`
  * Library: `/api/v1/music/library/*` routes
  * Player state: `GET/PUT /api/v1/music/player-state/`
* Key files involved (list with paths):
  * `backend/internal/api/v1/music/handler.go`
  * `backend/internal/api/v1/music/service.go`
  * `backend/internal/api/v1/music/catalog_service.go`
  * `backend/internal/api/v1/music/repository.go`
  * `frontend/src/service/music.ts`
  * `frontend/src/service/playlist.ts`
  * `frontend/src/service/playerState.ts`
  * `frontend/src/components/providers/GlobalMusicProvider.tsx`
  * `frontend/src/components/providers/musicProvider/musicProvider.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend providers + audio engine hooks.
  * Backend handler/service/repository with catalog-focused service split.
* Data flow (step-by-step):
  * Catalog endpoints query audio metadata joined with `home_file`.
  * Playlist commands update `playlist` and `playlist_track` with order management.
  * `player_state` persists current file, playlist, position, volume, shuffle/repeat.
  * Frontend global provider hydrates playback queue and updates backend periodically.
* External integrations:
  * Browser audio/media session APIs on frontend.
* State management (if applicable):
  * Global music React context + React Query for playlist/library data.

## 4. Data Model

* Entities involved:
  * Music track (file + audio metadata), playlist, playlist track, player state.
* Database tables / schemas:
  * `home_file`
  * `audio_metadata`
  * `playlist`
  * `playlist_track`
  * `player_state`
* Relationships:
  * `playlist_track.playlist_id -> playlist.id`
  * `playlist_track.file_id -> home_file.id`
  * `player_state.current_file_id -> home_file.id`
* Important fields:
  * `playlist.is_system`
  * `playlist_track.position`
  * `player_state.client_id`, `current_position`, `repeat_mode`, `shuffle`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Auto playlists exist and are read-only for edit operations.
  * Now Playing playlist is created if missing.
  * Track reorder updates positional sequence.
* Edge cases handled:
  * Missing playlist/track returns not found.
  * Duplicate track addition is prevented by unique constraint.
* Validation logic:
  * DTO parsing and service checks around playlist mutability and track membership.

## 6. User Flows

* Normal flow:
  * User opens Music -> navigates catalog -> starts playback -> queue state persists.
  * User creates playlist -> adds/removes/reorders tracks.
* Error flow:
  * Updating auto playlist returns read-only/business-rule error.
  * Missing file references or invalid IDs return 4xx/5xx.
* Edge cases:
  * Player state may diverge when multiple clients share same IP-based identity.

## 7. API / Interfaces

* Endpoints:
  * `/api/v1/music/playlists*`
  * `/api/v1/music/library*`
  * `/api/v1/music/player-state/`
* Input/output:
  * Playlist CRUD payloads, reorder payloads, paginated track/library outputs, player state DTOs.
* Contracts:
  * Frontend depends on stable negative IDs for automatic playlists and queue hydration shape.
* Internal interfaces:
  * `music.ServiceInterface`
  * `music.RepositoryInterface`

## 8. Problems & Limitations

* Technical debt:
  * Library/catalog assembly logic is extensive and can be hard to evolve.
* Bugs or inconsistencies:
  * State identity by IP is fragile for NAT/shared networks.
* Performance issues:
  * Large music libraries can increase catalog aggregation cost.
* Missing validations:
  * No authentication-scoped ownership of playlists/player state.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscation detected.
* External code execution:
  * None inside music module itself.
* Unsafe patterns:
  * IP-based `client_id` can leak/control another client session in shared network scenarios.
* Injection risks:
  * SQL repository pattern appears parameterized; risk remains if future query composition regresses.
* Hardcoded secrets:
  * None detected.
* Unsafe file/system access:
  * Playback depends on file stream endpoints from files module; no extra direct FS writes here.

## 10. Improvement Opportunities

* Refactors:
  * Separate playlist management from catalog query service into clearer bounded contexts.
* Architecture improvements:
  * Introduce authenticated user identity for player/playlist state.
* Scalability:
  * Materialize artist/album/genre summaries for large collections.
* UX improvements:
  * Improve conflict handling when queue differs across multiple active clients.

## 11. Acceptance Criteria

* Functional:
  * Users can browse music library, manage playlists, and resume playback state.
* Technical:
  * Auto playlists are protected from forbidden mutations.
  * Reorder and membership operations preserve deterministic track order.
* Edge cases:
  * Invalid playlist/file IDs return deterministic errors.
  * State persistence survives page reload and app restart.

## 12. Open Questions

* Unknown behaviors:
  * Expected reconciliation strategy for simultaneous updates from multiple clients.
* Missing clarity in code:
  * No explicit retention/cleanup policy for stale `player_state` rows.
* Assumptions made:
  * Music playback is intended for trusted household/local network usage.
