## 1. Feature Overview
- Name: Image Gallery and Classification
- Summary: Image browsing experience with grouping, album/folder collections, viewer tools, and metadata-based classification (heuristic + optional AI).
- Target user: User managing photo/screenshot/document image libraries.
- Business value: Converts raw image files into navigable visual collections.

## 2. Problem Statement
- Flat file browsing is poor for large image libraries.
- Without grouping/classification, image retrieval and curation are slow.

## 3. Current Behavior
- Frontend supports sections: library, recent, captures, photos, folders, albums.
- Grouping modes: date, type, name.
- Viewer supports zoom, filmstrip, slideshow interval from settings, and detail panel.
- Backend classification:
  - Heuristic capture/photo/other.
  - AI fallback when heuristic confidence is below threshold and AI service exists.
  - Stored classification category/confidence in image metadata.

## 4. User Flows
### 4.1 Main Flow
1. User opens `/images`.
2. System fetches paginated image dataset with metadata.
3. User navigates section/folder/album and opens viewer.
4. User can slide through images, zoom, and run slideshow.

### 4.2 Alternative Flows
1. User opens image result from global search with `image` and `imagePath` params; provider resolves missing item by path.
2. User toggles image favorites from image content actions.

### 4.3 Error Scenarios
1. Invalid `group_by` parameter is rejected by backend.
2. AI classification failure falls back to heuristic classification.
3. Missing metadata reduces classification fidelity and collection quality.

## 5. Functional Requirements
- The system must list image files with metadata and classification.
- The user can switch sections and grouping modes.
- The system must support folder and automatic album collections.
- The system should classify images and persist category/confidence.

## 6. Business Rules
- AI classification runs only when heuristic confidence is below threshold.
- Allowed AI categories are constrained to known category list.
- Capture detection prioritizes filename/path/software keyword hints.
- Photo confidence increases with camera EXIF evidence and path hints.

## 7. Data Model (Business View)
- File (image-type record in `home_file`).
- Image Metadata (`image_metadata`): EXIF/technical fields.
- Image Classification: category + confidence columns.
- Frontend derived collections: folder collections and automatic albums.

## 8. Interfaces
- User interfaces: `/images/*`, image domain header/sidebar, image viewer modal.
- APIs:
  - `GET /files/images?page&page_size&group_by`
  - `GET /files/path` (search deep-link resolution)
  - `POST /files/starred/:id`

## 9. Dependencies
- Files domain (source data, starring).
- Settings (`image_slideshow_seconds`).
- AI service (optional enhancement).

## 10. Limitations / Gaps
- Persisted image category type in frontend types is narrower than backend AI category vocabulary.
- Collection taxonomy is heuristic and may misclassify edge content.

## 11. Opportunities
- Add user override for category and feedback loop for model tuning.
- Add face/person and location-based collections.

## 12. Acceptance Criteria
- Given images exist, when user loads images page, then grouped image collections are displayed.
- Given low-confidence heuristic classification and AI enabled, when metadata is processed, then AI classification may override fallback.
- Given invalid `group_by`, when request is sent, then backend returns validation error.

## 13. Assumptions
- Metadata extraction pipeline has already processed image files for best experience.
- Albums are derived views, not persisted standalone entities.
