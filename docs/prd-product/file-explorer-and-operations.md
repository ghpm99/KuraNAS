## 1. Feature Overview
- Name: File Explorer and File Operations
- Summary: Core NAS browsing and filesystem management (tree navigation, upload, create folder, move, copy, rename, delete, streaming, thumbnails, metrics).
- Target user: User managing local/NAS file hierarchy.
- Business value: Delivers the primary storage-management capability of the product.

## 2. Problem Statement
- Users need to browse and operate on files/directories inside a controlled root path.
- Without this feature, KuraNAS cannot act as a usable NAS manager.

## 3. Current Behavior
- Tree browsing is scoped to configured entry point and supports categories (`all`, `starred`, `recent`).
- File operations validate IDs/paths and enforce root-bound path resolution.
- Upload writes files to disk first, then creates background processing job with dependent steps.
- Streaming endpoints support HTTP range for audio/video.
- Thumbnail/preview generation supports cached render and fallback icons.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/files`.
2. UI resolves selected path from URL and loads current node children via `/files/tree`.
3. User performs operation (upload/create/move/copy/rename/delete).
4. Backend validates path within entry point, executes filesystem change, and triggers rescan job/task.
5. UI refetches tree and reflects updated state.

### 4.2 Alternative Flows
1. User opens file blob/download through `/files/blob/:id`.
2. User streams media through `/files/stream/:id` or `/files/video-stream/:id`.
3. User toggles favorite (`/files/starred/:id`) and sees category-driven list updates.

### 4.3 Error Scenarios
1. Target/source path outside entry point: request rejected.
2. Destination already exists: conflict error.
3. Invalid rename/folder name: bad request.
4. Missing file on disk during thumbnail read: item marked deleted (system path) and not found returned.

## 5. Functional Requirements
- The system must provide paginated tree navigation by selected parent and category.
- The user can upload multiple files to a target folder.
- The user can create folder, move, copy, rename, and delete files/folders.
- The system must stream audio/video with range support.
- The system should provide file metrics (space used, top files, duplicates, format distribution).

## 6. Business Rules
- All file operations are constrained to configured `ENTRY_POINT`.
- Deleting entry-point root is forbidden.
- Move/copy into itself is blocked for directories.
- Upload requires non-empty file list and unique destination names.
- Upload processing job creates step graph: persist -> metadata/checksum/(thumbnail)/(playlist_index for video).
- Favorites are boolean toggle on file record.

## 7. Data Model (Business View)
- File (`home_file`): path hierarchy, type, format, size, checksum, starred, deleted marker.
- Recent Access (`recent_file`): IP + file_id + timestamp.
- Metadata entities: image/audio/video metadata (1:1 by file path/file_id intent).
- Derived reports: duplicates, top-size list, size by format group.

## 8. Interfaces
- User interfaces: files explorer, file details, folder picker, file content/list/grid.
- APIs:
  - `GET /files/tree`, `GET /files/path`, `GET /files/:id`
  - `POST /files/upload`, `POST /files/folder`, `POST /files/move`, `POST /files/copy`, `POST /files/rename`, `DELETE /files/path`
  - `POST /files/starred/:id`, `POST /files/update`
  - `GET /files/blob/:id`, `GET /files/thumbnail/:id`, `GET /files/video-thumbnail/:id`, `GET /files/video-preview/:id`
  - `GET /files/stream/:id`, `GET /files/video-stream/:id`
  - `GET /files/total-space-used`, `GET /files/top-files-by-size`, `GET /files/duplicate-files`

## 9. Dependencies
- Jobs/workers for post-upload processing and rescans.
- Metadata extraction scripts (Python) and ffmpeg for video previews.
- Configuration entry point and i18n messages.

## 10. Limitations / Gaps
- No user identity; recents and player-related data are IP-based.
- Full file blob reads entire file into memory.
- Some operation responses are generic success flags with limited domain context.

## 11. Opportunities
- Add optimistic UI with rollback for file operations.
- Add richer conflict resolution (rename on collision).
- Add chunked upload/resume and large-file upload progress.

## 12. Acceptance Criteria
- Given a valid folder, when user uploads files, then files are persisted and upload job_id is returned.
- Given source and destination valid, when move/copy executes, then source/destination scans are enqueued.
- Given media file exists, when range header is sent, then endpoint returns partial content and valid `Content-Range`.
- Given path escapes entry point, when operation is attempted, then request is rejected.

## 13. Assumptions
- Entry point is configured and accessible by backend process.
- Filesystem is authoritative; DB is synchronized via scan/job pipeline.
