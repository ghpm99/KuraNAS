## 1. Feature Overview
- Name: Global Search
- Summary: Unified cross-domain search dialog combining quick actions and indexed search results (files, folders, artists, albums, playlists, videos, images).
- Target user: User navigating quickly across large libraries.
- Business value: Reduces time-to-content and improves discoverability.

## 2. Problem Statement
- Domain-specific search is inefficient for mixed media libraries.
- Without global search, users must context-switch across pages.

## 3. Current Behavior
- Search dialog opens via keyboard shortcut (`Ctrl/Cmd+K`).
- Query threshold for API search is >= 2 characters.
- Backend runs multi-entity SQL search with per-section limits.
- Optional AI expands query keywords and suggestion for broader recall.
- Results map to direct navigation actions by entity type.

## 4. User Flows
### 4.1 Main Flow
1. User presses shortcut or triggers search UI.
2. User types query.
3. System fetches `/search/global?q&limit` and merges sections.
4. User navigates with arrows/enter or mouse to open result target.

### 4.2 Alternative Flows
1. Query matches quick actions only and no indexed entities.
2. AI expansion adds additional results for semantically related keywords.

### 4.3 Error Scenarios
1. Invalid limit query returns bad request.
2. Search backend failure returns generic error payload.
3. AI expansion/parsing failure silently falls back to base SQL search.

## 5. Functional Requirements
- The system must provide a single endpoint for cross-domain search.
- The user can open/close search globally and keyboard-navigate results.
- The system should include quick navigation actions even without query hits.
- The system should enrich search with AI keyword expansion when available.

## 6. Business Rules
- Default limit is 6; max limit is 12.
- Empty query returns empty entity arrays.
- AI expansion runs only for queries with at least 2 words.
- Duplicate entity IDs are deduplicated when merging AI-expanded results.

## 7. Data Model (Business View)
- Aggregated search response with typed sections:
  - files, folders, artists, albums, playlists (music/video), videos, images.
- Result entities include normalized routing identifiers (id/key/path/classification).

## 8. Interfaces
- User interfaces: global search dialog and provider.
- API: `GET /search/global?q=<query>&limit=<n>`.

## 9. Dependencies
- Indexed file/music/video/image/playlist data.
- AI service for query expansion (optional).
- Routing helpers for deep linking into domains.

## 10. Limitations / Gaps
- Result ranking is primarily SQL/section order; no unified relevance scoring across domains.
- Suggestions from backend are currently not surfaced in frontend UI.

## 11. Opportunities
- Introduce cross-domain relevance ranking.
- Show search suggestion/intent disambiguation in dialog.

## 12. Acceptance Criteria
- Given query length >= 2, when user searches, then dialog shows sectioned results from `/search/global`.
- Given empty query, when search executes, then no entity results are returned.
- Given AI is unavailable, when query executes, then base search still returns normal results.

## 13. Assumptions
- Search scope is limited to indexed records, not raw filesystem walk on demand.
