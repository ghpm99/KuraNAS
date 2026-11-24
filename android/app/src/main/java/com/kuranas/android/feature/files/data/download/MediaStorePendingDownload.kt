package com.kuranas.android.feature.files.data.download

import android.content.ContentResolver
import android.content.ContentValues
import android.net.Uri
import android.provider.MediaStore
import java.io.FileNotFoundException
import java.io.IOException
import java.io.OutputStream

class MediaStorePendingDownload(
    private val contentResolver: ContentResolver,
    private val fileName: String,
    private val mimeType: String,
    private var entryUri: Uri? = null,
) : PendingDownload {

    private val collection = downloadCollectionFor(mimeType)

    override val reference: String?
        get() = entryUri?.toString()

    override fun openOutput(shouldAppend: Boolean): OutputStream {
        val uri = entryUri ?: insertPendingEntry().also { entryUri = it }
        val mode = if (shouldAppend) APPEND_MODE else TRUNCATE_MODE
        return contentResolver.openOutputStream(uri, mode) ?: throw IOException("No output stream for $uri")
    }

    override fun commit(): Uri {
        val uri = entryUri ?: throw IOException("Nothing was written")
        val finishedValues = ContentValues().apply { put(MediaStore.MediaColumns.IS_PENDING, 0) }
        contentResolver.update(uri, finishedValues, null, null)
        return uri
    }

    override fun discard() {
        val uri = entryUri ?: return
        entryUri = null
        contentResolver.delete(uri, null, null)
    }

    fun writtenBytes(): Long? = try {
        val uri = entryUri ?: return null
        contentResolver.openFileDescriptor(uri, "r")?.use { it.statSize }
    } catch (missing: FileNotFoundException) {
        null
    } catch (denied: SecurityException) {
        null
    }

    private fun insertPendingEntry(): Uri {
        val values = ContentValues().apply {
            put(MediaStore.MediaColumns.DISPLAY_NAME, fileName)
            put(MediaStore.MediaColumns.MIME_TYPE, mimeType)
            put(MediaStore.MediaColumns.RELATIVE_PATH, collection.relativeDirectory)
            put(MediaStore.MediaColumns.IS_PENDING, 1)
        }
        return contentResolver.insert(collection.contentUri(), values) ?: throw IOException("Could not create $fileName")
    }

    private fun DownloadCollection.contentUri(): Uri = when (this) {
        DownloadCollection.IMAGES -> MediaStore.Images.Media.EXTERNAL_CONTENT_URI
        DownloadCollection.VIDEOS -> MediaStore.Video.Media.EXTERNAL_CONTENT_URI
        DownloadCollection.DOWNLOADS -> MediaStore.Downloads.EXTERNAL_CONTENT_URI
    }

    private companion object {
        const val APPEND_MODE = "wa"
        const val TRUNCATE_MODE = "wt"
    }
}
