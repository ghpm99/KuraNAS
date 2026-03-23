## 1. Feature Overview
- Name: Video Library, Playback, and Smart Playlists
- Summary: Video experience spanning home catalog sections, playlist-based playback sessions, behavior tracking, context playlists, and smart playlist generation.
- Target user: User consuming mixed personal and catalog-style video collections.
- Business value: Enables continuous video consumption and auto-organization at scale.

## 2. Problem Statement
- Raw video file lists are not enough for browse-and-watch workflows.
- Without session state and smart grouping, discovery and continuation are weak.

## 3. Current Behavior
- Playback session is playlist-centric and persisted per client IP.
- Start playback can use explicit playlist or auto-resolved folder-context playlist.
- Behavior events recorded: started, paused, resumed, completed, skipped, abandoned.
- Home catalog sections: continue, series, movies, personal, recent.
- Smart playlists rebuilt from metadata + behavior scoring engine and exclusions.
- Auto playlist removal creates exclusion record so rebuild does not re-add excluded item.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/videos` (sectioned home).
2. User selects playlist/video; app routes to `/video/:id` with context params.
3. Backend starts playback session and returns playlist + state.
4. Player sends periodic state updates and completion signal.
5. User navigates next/previous inside playlist session.

### 4.2 Alternative Flows
1. User browses folder section and adds library videos into a selected playlist.
2. User edits playlist name/order and removes items from detail view.
3. Background worker triggers smart playlist rebuild after file processing.

### 4.3 Error Scenarios
1. Video not in selected playlist: bad request.
2. Navigation without valid playback state/playlist: bad request.
3. Invalid behavior event type: bad request.
4. Context path has no videos: not found.

## 5. Functional Requirements
- The system must start/track/update video playback sessions.
- The user can navigate next/previous videos within active playlist context.
- The system must expose home catalog sections and playlist detail views.
- The user can hide playlist, rename playlist, reorder items, and add/remove videos.
- The system should generate smart playlists from metadata and behavior data.

## 6. Business Rules
- Playback and behavior tracking are keyed by client IP.
- Progress status values are `not_started`, `in_progress`, `completed`.
- Reorder payload cannot contain duplicate `video_id` or duplicate `order_index`.
- Empty playlist rename value is invalid.
- Smart engine applies deduplication and negative-score filtering before persistence.
- Hidden playlists are excluded unless `include_hidden=true`.

## 7. Data Model (Business View)
- Video Playlist (`video_playlist`): type/source/classification/visibility and play metadata.
- Playlist Items (`video_playlist_item`): ordered videos with source kind (`auto`/`manual`).
- Playback State (`video_playback_state`): current client playback context.
- Behavior Event (`video_behavior_event`): granular viewing behavior for scoring/progress.
- Playlist Exclusion (`video_playlist_exclusion`): explicit suppression list for auto playlists.

## 8. Interfaces
- User interfaces: `/videos/*`, video section grids, playlist detail, dedicated `/video/:id` player.
- APIs:
  - Playback: `/video/playback/start|state|next|previous|behavior`
  - Catalog: `/video/catalog/home`
  - Library: `/video/library/files`
  - Playlists: `/video/playlists/*` (list/detail/memberships/rebuild/unassigned/reorder/hide/add/remove/update)

## 9. Dependencies
- Files/video metadata extraction pipeline.
- Worker jobs (playlist_index step and rebuild orchestration).
- Optional AI for catalog section descriptions.

## 10. Limitations / Gaps
- Per-IP state can collide on shared networks.
- Section titles in backend catalog are currently hardcoded strings.
- Autoplay behavior depends on frontend settings only.

## 11. Opportunities
- Profile-based watch history.
- Better conflict-safe playlist slug routing (ID-based fallback).
- Advanced recommendation rails from behavior model.

## 12. Acceptance Criteria
- Given valid video and playlist, when start playback is requested, then session returns playlist + playback state.
- Given playback completes, when completed update is sent, then completion event is stored and progress becomes completed.
- Given auto playlist item is removed, when rebuild runs, then excluded video is not reinserted.
- Given reorder request has duplicate order indexes, when processed, then request is rejected.

## 13. Assumptions
- Video metadata and behavior data are sufficient inputs for smart grouping quality.
