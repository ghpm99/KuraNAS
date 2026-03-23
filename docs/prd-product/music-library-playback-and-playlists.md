## 1. Feature Overview
- Name: Music Library, Playback State, and Playlists
- Summary: Music domain covering catalog browsing (artists/albums/genres/folders), manual playlists, automatic playlists, now playing queue, and persisted player state.
- Target user: User consuming personal music library with queue/playlist workflows.
- Business value: Turns file-based music into a usable media experience with continuity.

## 2. Problem Statement
- Users need structure and playback continuity beyond raw file listings.
- Without this feature, music consumption is fragmented and non-persistent.

## 3. Current Behavior
- Library catalog is built from indexed audio metadata and normalized keys.
- Playlists include manual CRUD plus system/automatic playlists (negative IDs).
- Automatic playlists currently include Continue Listening, Recently Added, Favorites.
- Player state is stored/retrieved by client IP (`client_id`).
- Global music player syncs queue/progress/controls to backend and can rehydrate queue.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/music` and selects a section (home/playlists/artists/albums/genres/folders).
2. System loads catalog or playlist data.
3. User starts playback from a track list/playlist.
4. Global player updates local audio engine and debounced backend player-state sync.

### 4.2 Alternative Flows
1. User opens now-playing queue (`/music/playlists/now-playing`) and resumes persisted file.
2. User creates/deletes playlists and adds/removes tracks.
3. User uses home featured cards to replace queue by artist/album/playlist context.

### 4.3 Error Scenarios
1. Attempt to mutate automatic playlist: rejected (read-only).
2. Playlist or tracks not found: not found response.
3. Invalid payload for playlist/player-state mutation: bad request.

## 5. Functional Requirements
- The system must provide grouped library views and track lists by key.
- The user can create, update, delete manual playlists.
- The user can add/remove/reorder tracks in manual playlists.
- The system must provide automatic playlists and now playing queue.
- The system must persist and restore playback state.

## 6. Business Rules
- Automatic playlists use reserved negative IDs.
- Automatic playlists are immutable through CRUD/track mutation endpoints.
- Genre normalization consolidates synonymous labels (e.g., hip-hop variants).
- Continue Listening seeds from current file in player state, then current playlist tracks, then recent interactions.
- Player state fields include volume, shuffle, repeat mode, current position, current file, playlist.

## 7. Data Model (Business View)
- Playlist (`playlist`): manual/system queue records.
- Playlist Track (`playlist_track`): ordered relation playlist -> file.
- Player State (`player_state`): one record per client IP.
- Music Index Entry: derived join of file + audio metadata for grouping.

## 8. Interfaces
- User interfaces: music home, playlist list/detail, artist/album/genre/folder views, global player, queue drawer.
- APIs:
  - `GET/POST/PUT/DELETE /music/playlists/*`
  - `GET/PUT /music/player-state/`
  - `GET /music/library/*` (home, grouped catalogs, track lists)

## 9. Dependencies
- Files and audio metadata extraction pipeline.
- Settings (`remember_music_queue`) for hydration behavior.
- i18n for playlist labels and UI strings.

## 10. Limitations / Gaps
- Identity model is IP-based, not authenticated user-based.
- System playlist semantics are implicit by ID range and flags.

## 11. Opportunities
- Multi-device/user profile playback sync.
- Smart recommendations based on listening events.
- Rich playlist operations (bulk edit, smart rules, collaborative sets).

## 12. Acceptance Criteria
- Given a manual playlist, when user adds track, then track appears in playlist detail and count updates.
- Given an automatic playlist ID, when user tries to update/delete it, then request is rejected.
- Given player state exists, when user reopens app, then current queue/position can be resumed.
- Given grouped catalog request, when key is provided, then system returns matching track subset with deterministic ordering.

## 13. Assumptions
- Audio metadata exists for best grouping quality; missing metadata still allows file-level playback.
