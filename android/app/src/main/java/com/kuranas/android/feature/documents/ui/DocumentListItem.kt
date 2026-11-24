package com.kuranas.android.feature.documents.ui

import com.kuranas.android.feature.documents.data.DocumentDto
import com.kuranas.android.feature.documents.data.DocumentSearchResultDto

data class DocumentListItem(
    val id: Int,
    val name: String,
    val format: String,
    val parentPath: String,
    val size: Long,
    val timestamp: String,
    val isStarred: Boolean,
    val snippet: String? = null,
    val pageNumber: Int? = null,
)

internal fun DocumentDto.toListItem() = DocumentListItem(
    id = id,
    name = name,
    format = format,
    parentPath = parentPath,
    size = size,
    timestamp = modifiedAt,
    isStarred = starred,
)

internal fun DocumentSearchResultDto.toListItem() = DocumentListItem(
    id = fileId,
    name = name,
    format = format,
    parentPath = parentPath,
    size = size,
    timestamp = updatedAt,
    isStarred = starred,
    snippet = snippet.takeIf { it.isNotBlank() },
    pageNumber = page.takeIf { it > 0 },
)
