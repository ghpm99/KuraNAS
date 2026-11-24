package com.kuranas.android.feature.cameraupload.data

import android.content.ContentResolver
import android.content.ContentUris
import android.net.Uri
import android.provider.MediaStore
import android.provider.MediaStore.MediaColumns
import dagger.Reusable
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject

@Reusable
class CameraMediaScanner @Inject constructor(private val contentResolver: ContentResolver) {

    /**
     * Lista a mídia da câmera com DATE_ADDED (em segundos) maior que [addedAfterSeconds].
     * Lança [SecurityException] quando falta a permissão de leitura de mídia.
     */
    suspend fun scan(includeVideos: Boolean, addedAfterSeconds: Long): List<CameraMedia> = withContext(Dispatchers.IO) {
        val query = cameraMediaQuery(addedAfterSeconds)
        val images = queryKind(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, CameraMediaKind.IMAGE, query)
        val videos = if (includeVideos) queryKind(MediaStore.Video.Media.EXTERNAL_CONTENT_URI, CameraMediaKind.VIDEO, query) else emptyList()
        images + videos
    }

    private fun queryKind(collectionUri: Uri, kind: CameraMediaKind, query: CameraMediaQuery): List<CameraMedia> {
        val cursor = contentResolver.query(collectionUri, PROJECTION, query.selection, query.selectionArguments, null)
            ?: return emptyList()
        return cursor.use {
            val idColumn = it.getColumnIndexOrThrow(MediaColumns._ID)
            val nameColumn = it.getColumnIndexOrThrow(MediaColumns.DISPLAY_NAME)
            val sizeColumn = it.getColumnIndexOrThrow(MediaColumns.SIZE)
            val takenColumn = it.getColumnIndexOrThrow(MediaColumns.DATE_TAKEN)
            val addedColumn = it.getColumnIndexOrThrow(MediaColumns.DATE_ADDED)
            val modifiedColumn = it.getColumnIndexOrThrow(MediaColumns.DATE_MODIFIED)
            buildList {
                while (it.moveToNext()) {
                    val dateTakenMillis = it.getLong(takenColumn).takeIf { taken -> taken > 0 }
                        ?: (it.getLong(addedColumn) * MILLIS_PER_SECOND)
                    add(
                        CameraMedia(
                            id = it.getLong(idColumn),
                            kind = kind,
                            contentUri = ContentUris.withAppendedId(collectionUri, it.getLong(idColumn)).toString(),
                            displayName = it.getString(nameColumn).orEmpty(),
                            sizeBytes = it.getLong(sizeColumn),
                            dateTakenMillis = dateTakenMillis,
                            dateModifiedMillis = it.getLong(modifiedColumn) * MILLIS_PER_SECOND,
                        ),
                    )
                }
            }
        }
    }

    private companion object {
        const val MILLIS_PER_SECOND = 1000L
        val PROJECTION = arrayOf(
            MediaColumns._ID,
            MediaColumns.DISPLAY_NAME,
            MediaColumns.SIZE,
            MediaColumns.DATE_TAKEN,
            MediaColumns.DATE_ADDED,
            MediaColumns.DATE_MODIFIED,
        )
    }
}
