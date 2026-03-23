## 1. Feature Overview

* Name: Global Search
* Summary: Unified search across files, folders, music entities, video playlists/videos, and images, with optional AI query expansion.
* Purpose: Provide a single entry point to navigate the full content graph quickly.
* Business value: Reduces navigation friction and increases retrieval success in large libraries.

## 2. Current Implementation

* How it works today: `GET /search/global` executes multi-domain SQL queries and optionally augments results using AI-generated keywords.
* Main flows:
  * User types query in global search dialog.
  * Backend normalizes query and enforces result limit.
  * Repository executes per-domain search (files/folders/artists/albums/playlists/videos/images).
  * If AI is enabled and query has >=2 words, AI extraction returns keywords/suggestion, then backend merges extra results by deduping ids.
* Entry points (routes, handlers, jobs):
  * `GET /api/v1/search/global?q=&limit=`
* Key files involved (list with paths):
  * `backend/internal/api/v1/search/handler.go`
  * `backend/internal/api/v1/search/service.go`
  * `backend/internal/api/v1/search/repository.go`
  * `backend/internal/api/v1/search/dto.go`
  * `frontend/src/service/search.ts`
  * `frontend/src/components/search/GlobalSearchProvider.tsx`
  * `frontend/src/components/search/GlobalSearchDialog.tsx`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Frontend dialog/provider with keyboard-first interaction.
  * Backend handler -> service -> repository plus optional AI service.
* Data flow (step-by-step):
  * Query and limit parsed.
  * Service executes baseline multi-repository search.
  * AI expansion prompt can extract keywords and suggestion text.
  * Service reruns selected searches for each keyword and merges unique results.
  * Response returned as typed grouped payload.
* External integrations:
  * AI provider via internal AI router.
* State management (if applicable):
  * Global search provider state for dialog open/close, query text, and results.

## 4. Data Model

* Entities involved:
  * Search result DTO groups per domain.
* Database tables / schemas:
  * Uses indexes/data across `home_file`, metadata tables, `playlist`, `video_playlist` and related item tables.
* Relationships:
  * Playlist and media joins are used for richer scoped results.
* Important fields:
  * Search `query`, `suggestion`, grouped arrays with ids/keys/path metadata.

## 5. Business Rules

* Explicit rules implemented in code:
  * Default limit is 6; maximum is 12.
  * Empty query returns empty grouped response.
  * AI expansion is skipped for single-word queries and when AI service is unavailable.
* Edge cases handled:
  * AI JSON parse failures are tolerated; baseline search still returns.
  * Duplicate elimination by id during merge.
* Validation logic:
  * Limit clamping and query trimming.

## 6. User Flows

* Normal flow:
  * User opens global search -> types query -> receives grouped results -> navigates directly to selected item.
* Error flow:
  * Backend search failure returns server error; dialog should handle error state.
* Edge cases:
  * AI expansion may add noisy keywords; merged results can exceed initial relevance intent.

## 7. API / Interfaces

* Endpoints:
  * `GET /api/v1/search/global`
* Input/output:
  * Query params `q`, `limit`; response with grouped result arrays and optional suggestion.
* Contracts:
  * Frontend expects stable group keys and DTO fields per scope.
* Internal interfaces:
  * `search.ServiceInterface`
  * `search.RepositoryInterface`

## 8. Problems & Limitations

* Technical debt:
  * Single endpoint combines many concerns (query parsing, ranking, AI expansion).
* Bugs or inconsistencies:
  * Relevance ranking is mostly DB-order/append order; limited scoring normalization across domains.
* Performance issues:
  * Multiple domain queries + keyword expansion can multiply DB load.
* Missing validations:
  * No explicit abuse/rate limiting for rapid search bursts.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated behavior found.
* External code execution:
  * No shell execution; optional external AI API calls present.
* Unsafe patterns:
  * AI-generated keywords are fed back into search; should remain bounded and audited for prompt-injection side effects.
* Injection risks:
  * Query string is user-controlled; repository must stay parameterized and avoid string-concatenated SQL.
* Hardcoded secrets:
  * None in search module.
* Unsafe file/system access:
  * None direct.

## 10. Improvement Opportunities

* Refactors:
  * Introduce ranking abstraction to normalize relevance across domains.
* Architecture improvements:
  * Add search indexing/search-service layer for scalable cross-domain retrieval.
* Scalability:
  * Cache frequent query results and debounce frontend requests.
* UX improvements:
  * Add highlighted matches and result-type filters.

## 11. Acceptance Criteria

* Functional:
  * Search returns grouped results across supported domains for non-empty queries.
* Technical:
  * Limit clamping and empty-query behavior are deterministic.
  * AI expansion remains optional and non-breaking when unavailable.
* Edge cases:
  * Parse/AI failures still return baseline results when possible.

## 12. Open Questions

* Unknown behaviors:
  * Desired relevance strategy between file name match vs metadata match.
* Missing clarity in code:
  * No explicit telemetry for search quality/CTR.
* Assumptions made:
  * Search is expected to run against moderate dataset sizes without dedicated search engine.
