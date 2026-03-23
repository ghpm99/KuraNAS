## 1. Feature Overview

* Name: Video Playback and Smart Playlists
* Summary: Provides video playback session control, behavior tracking, home catalog sections, and automatic playlist generation/rebuild.
* Purpose: Organize large video libraries into watchable contexts (continue watching, series, movies, personal, etc.).
* Business value: Improves content discovery and retention via recommendation-like sequencing and progress continuity.

## 2. Current Implementation

* How it works today: Video module persists playback state/events, serves catalog/playlists, and rebuilds smart playlists through a rule/scoring engine.
* Main flows:
  * Player starts/updates session and records behavior events.
  * Catalog home endpoint aggregates sectioned playlists/items.
  * Rebuild endpoint runs classifier + strategy chain + scorer and persists auto playlists/items.
  * Manual actions include hide/unhide, add/remove membership, reorder, update playlist metadata.
* Entry points (routes, handlers, jobs):
  * Playback: `POST /api/v1/video/playback/start`, `GET/PUT /state`, `POST /next`, `POST /previous`, `POST /behavior`
  * Catalog: `GET /api/v1/video/catalog/home`
  * Library: `GET /api/v1/video/library/files`
  * Playlists: `GET/PUT /api/v1/video/playlists*`, `POST /rebuild`, `GET /unassigned`, membership routes
* Key files involved (list with paths):
  * `backend/internal/api/v1/video/handler.go`
  * `backend/internal/api/v1/video/service.go`
  * `backend/internal/api/v1/video/repository.go`
  * `backend/internal/api/v1/video/playlist/engine.go`
  * `backend/internal/api/v1/video/playlist/classifier.go`
  * `backend/internal/api/v1/video/playlist/strategy.go`
  * `backend/internal/api/v1/video/playlist/scorer.go`
  * `frontend/src/service/videoPlayback.ts`
  * `frontend/src/components/providers/videoContentProvider/videoContentProvider.tsx`
  * `frontend/src/pages/videos/videos.tsx`
  * `frontend/src/pages/videoPlayer/videoPlayer.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend video content provider + player page/hooks.
  * Backend handler/service/repository + dedicated playlist engine package.
* Data flow (step-by-step):
  * Playback endpoints read/write `video_playback_state` and append `video_behavior_event`.
  * Catalog request loads playlists/items and composes UI sections.
  * Rebuild loads candidate videos + metadata + behavior, classifies items, computes strategies, scores candidates, deduplicates, and stores auto playlists.
  * Frontend consumes sectioned DTOs and drives player navigation (next/previous/resume).
* External integrations:
  * Optional AI generation of section descriptions.
* State management (if applicable):
  * Video provider context + React Query for catalog/playback data.

## 4. Data Model

* Entities involved:
  * Video playlist, playlist item, exclusion, playback state, behavior event, video file + metadata.
* Database tables / schemas:
  * `video_playlist`
  * `video_playlist_item`
  * `video_playlist_exclusion`
  * `video_playback_state`
  * `video_behavior_event`
  * `home_file`
  * `video_metadata`
* Relationships:
  * `video_playlist_item.playlist_id -> video_playlist.id`
  * `video_playlist_item.video_id -> home_file.id`
  * `video_behavior_event.video_id -> home_file.id`
  * Playback state references playlist/video ids.
* Important fields:
  * `video_playlist.type`, `classification`, `group_mode`, `is_auto`, `is_hidden`.
  * `video_behavior_event.event_type`, `watched_pct`, `position`, `duration`.

## 5. Business Rules

* Explicit rules implemented in code:
  * Playlist type constraints include `folder`, `series`, `movie`, `custom`, `course`, `mixed`, `continue`.
  * Rebuild process updates auto playlists and preserves manual exclusions.
  * Behavior events influence continue/resume and recommendation ranking.
* Edge cases handled:
  * Missing playback session returns defaults/fallback behavior.
  * Unassigned videos endpoint exposes items not currently grouped.
* Validation logic:
  * DTO validation for playlist/playback updates and reorder operations.

## 6. User Flows

* Normal flow:
  * User opens Videos home -> chooses section/playlist -> starts playback -> session progress saved -> returns to continue watching.
* Error flow:
  * Invalid playlist/video IDs or state payloads return errors.
  * Rebuild failure keeps previous playlist state.
* Edge cases:
  * Multi-client same-IP playback can overwrite state.
  * Rebuild heuristics may classify ambiguous content inconsistently.

## 7. API / Interfaces

* Endpoints:
  * `/api/v1/video/playback/*`
  * `/api/v1/video/catalog/home`
  * `/api/v1/video/library/files`
  * `/api/v1/video/playlists*`
* Input/output:
  * Playback control payloads, behavior event payloads, catalog/playlist DTOs.
* Contracts:
  * Frontend expects section keys and playlist item structures used by navigation/player controls.
* Internal interfaces:
  * `video.ServiceInterface`
  * `video.RepositoryInterface`
  * Playlist engine classifier/strategy/scorer interfaces and structs.

## 8. Problems & Limitations

* Technical debt:
  * Smart playlist engine contains many heuristics/rules without explicit tuning configuration.
* Bugs or inconsistencies:
  * Some section labels/descriptions are hardcoded in code paths, conflicting with strict i18n policy.
* Performance issues:
  * Full rebuild operations can be heavy for large libraries.
* Missing validations:
  * No explicit per-user authorization or ownership isolation for playback data.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code detected.
* External code execution:
  * None directly in video domain rebuild logic.
* Unsafe patterns:
  * IP-based `client_id` for playback and behavior tracking can cause cross-user state interference.
* Injection risks:
  * Search/filter inputs are mostly structured but should remain parameterized in repository queries.
* Hardcoded secrets:
  * None detected in this module.
* Unsafe file/system access:
  * Video playback relies on file streaming endpoints and metadata from filesystem-indexed paths.

## 10. Improvement Opportunities

* Refactors:
  * Move playlist engine constants/weights into explicit configuration.
* Architecture improvements:
  * Separate recommendation/classification pipeline from request path and run asynchronously.
* Scalability:
  * Incremental rebuild based on changed videos/events instead of full recomputation.
* UX improvements:
  * Expose explainability for why a video appears in a smart section.

## 11. Acceptance Criteria

* Functional:
  * Users can browse video sections, play videos, and resume from saved position.
  * Playlist rebuild produces valid auto playlists with stable ordering.
* Technical:
  * Playback state and behavior events persist and remain queryable.
  * Reorder/hide/membership endpoints update playlist state deterministically.
* Edge cases:
  * Invalid IDs and malformed payloads return clear failures.
  * Rebuild errors do not corrupt existing playlist records.

## 12. Open Questions

* Unknown behaviors:
  * Long-term behavior retention policy and decay logic for recommendations.
* Missing clarity in code:
  * No documented target quality metrics for classification/scoring outcomes.
* Assumptions made:
  * Smart playlist experience is optimized for a single trusted household context.
