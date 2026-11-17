package com.kuranas.android.feature.files.data

import android.content.ContentResolver
import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import com.kuranas.android.R

internal const val UNKNOWN_SIZE_BYTES = -1L

internal fun describeSelectedFile(context: Context, contentResolver: ContentResolver, uri: Uri): SelectedFile {
    val projection = arrayOf(OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE)
    contentResolver.query(uri, projection, null, null, null)?.use { cursor ->
        if (!cursor.moveToFirst()) return@use
        val nameColumn = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
        val sizeColumn = cursor.getColumnIndex(OpenableColumns.SIZE)
        val displayName = if (nameColumn >= 0) cursor.getString(nameColumn) else null
        val sizeBytes = if (sizeColumn >= 0 && !cursor.isNull(sizeColumn)) cursor.getLong(sizeColumn) else UNKNOWN_SIZE_BYTES
        if (!displayName.isNullOrBlank()) return SelectedFile(uri, displayName, sizeBytes)
        return SelectedFile(uri, fallbackFileName(context, uri), sizeBytes)
    }
    return SelectedFile(uri, fallbackFileName(context, uri), UNKNOWN_SIZE_BYTES)
}

private fun fallbackFileName(context: Context, uri: Uri): String =
    uri.lastPathSegment?.takeIf { it.isNotBlank() } ?: context.getString(R.string.upload_unnamed_file)
