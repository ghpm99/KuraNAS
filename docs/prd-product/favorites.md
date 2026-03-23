## 1. Feature Overview
- Name: Favorites (Starred Files)
- Summary: Specialized view over starred files with filters for folders, files, and media.
- Target user: User curating high-priority files for quick access.
- Business value: Increases retrieval speed for important items.

## 2. Problem Statement
- Users need a curated subset of files independent of deep folder location.
- Without favorites, repeated navigation to key files is inefficient.

## 3. Current Behavior
- Favorites page sets file-list filter to `starred` and reuses file tree context.
- Supports scoped filtering (`all`, `folders`, `files`, `media`) and list/grid view.
- Breadcrumb/path context still follows selected node in shared file tree.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/favorites`.
2. System loads starred tree data via file provider.
3. User applies local filter and opens item or navigates through breadcrumb.

### 4.2 Alternative Flows
1. User toggles star in files/home/other surfaces and favorites set updates after refetch.
2. User opens a selected file and sees details pane.

### 4.3 Error Scenarios
1. If provider data fails, screen cannot populate (depends on files domain state).
2. If selected node is not directory/file as expected, content panel may be empty.

## 5. Functional Requirements
- The system must show only starred items when favorites route is active.
- The user can switch between filter subsets and view modes.
- The system should preserve folder context and breadcrumbs.

## 6. Business Rules
- “Media” subset is based on inferred file type (`audio`, `image`, `video`) from extension.
- Favorites selection is global (single starred flag), not per user profile.

## 7. Data Model (Business View)
- Uses File entity with `starred` flag.
- Inherits file hierarchy and metadata from file explorer domain.

## 8. Interfaces
- User interfaces: `/favorites`, favorites screen cards/list, shared file details.
- APIs: indirect via file provider (`GET /files/tree?category=starred`, `POST /files/starred/:id`).

## 9. Dependencies
- File explorer/provider and file operations.
- i18n for all labels.

## 10. Limitations / Gaps
- No dedicated backend analytics for favorites behavior.
- Filtering is client-side on currently loaded scope.

## 11. Opportunities
- Add sorting presets (recently opened, size, type).
- Add favorite collections/tags.

## 12. Acceptance Criteria
- Given item is starred, when user opens favorites, then item appears in favorites dataset.
- Given user selects `media` filter, when list renders, then only media files are shown.
- Given user unstars an item, when data refreshes, then item is removed from favorites view.

## 13. Assumptions
- Favorites is intentionally an overlay of file explorer, not an independent storage model.
