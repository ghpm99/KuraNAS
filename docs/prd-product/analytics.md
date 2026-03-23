## 1. Feature Overview
- Name: Analytics Overview and Library Health
- Summary: Operational analytics panel covering storage KPIs, growth, distribution, duplicates, processing queue, and index health.
- Target user: User monitoring NAS health and library quality.
- Business value: Supports data-driven cleanup and indexing decisions.

## 2. Problem Statement
- Users need visibility into growth, utilization, indexing quality, and errors.
- Without analytics, maintenance is reactive and inefficient.

## 3. Current Behavior
- Backend serves consolidated overview data for periods (`24h`, `7d`, `30d`, `90d`).
- Frontend has two sections: overview and library.
- Metrics include storage totals, file/folder counts, type/extension breakdown, top folders, duplicates, recent files, processing backlog/failures, and health errors.
- Optional AI-generated textual insights are appended when AI service is enabled.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/analytics`.
2. System loads overview for selected period.
3. User switches between overview/library tabs and period values.
4. User reviews KPIs/tables and optionally navigates to files section actions.

### 4.2 Alternative Flows
1. User refreshes analytics manually from toolbar.
2. If AI is enabled, insights list is populated from summarized metrics.

### 4.3 Error Scenarios
1. Invalid period query returns bad request.
2. Data retrieval errors show analytics error state.
3. AI insights parsing failure yields empty insights without blocking analytics payload.

## 5. Functional Requirements
- The system must expose period-based analytics overview endpoint.
- The user can switch period and section views.
- The system should display index health and recent errors.
- The system should expose duplicates and reclaimable space indicators.

## 6. Business Rules
- Accepted periods: `24h`, `7d` (default), `30d`, `90d`.
- Health status mapping:
  - `PENDING` -> scanning
  - `FAILED` -> error
  - otherwise -> ok
- Processing failures aggregate metadata + thumbnail failures.

## 7. Data Model (Business View)
- Aggregated analytics dataset (no single table): storage kpis, time series, distributions, duplicate groups, library coverage, queue health, log errors.
- Relies on `home_file`, metadata tables, worker tables, and log table.

## 8. Interfaces
- User interfaces: analytics content with `overview` and `library` sub-views.
- API: `GET /analytics/overview?period=`.

## 9. Dependencies
- File index and metadata completeness.
- Worker pipeline statuses.
- Logs for recent error extraction.
- Optional AI service for insights.

## 10. Limitations / Gaps
- Several overview actions route generically to files instead of pre-filtered contexts.
- Analytics insight generation is best effort and unvalidated semantically.

## 11. Opportunities
- Add deep links with concrete filters (duplicates/top folders/errors).
- Add historical snapshots and trend comparisons per period.

## 12. Acceptance Criteria
- Given valid period, when overview is requested, then response contains storage, counts, distributions, health, and processing blocks.
- Given invalid period, when endpoint is called, then system returns bad request.
- Given AI unavailable, when overview loads, then insights array is empty and core analytics still render.

## 13. Assumptions
- Analytics is read-only and computed from existing operational data.
