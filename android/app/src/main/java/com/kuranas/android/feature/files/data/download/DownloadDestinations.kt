package com.kuranas.android.feature.files.data.download

import android.content.Context
import android.net.Uri
import androidx.core.content.FileProvider
import dagger.hilt.android.qualifiers.ApplicationContext
import java.io.File
import javax.inject.Inject

const val SHARED_CACHE_DIRECTORY = "shared"

fun fileProviderAuthority(context: Context): String = "${context.packageName}.fileprovider"

class MediaStoreDownloadDestinations @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    fun create(request: DownloadRequest, mimeType: String): PendingDownload =
        MediaStorePendingDownload(context.contentResolver, request.fileName, mimeType)
}

class CacheDownloadDestinations @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    fun create(request: DownloadRequest): PendingDownload {
        val fileFolder = File(File(context.cacheDir, SHARED_CACHE_DIRECTORY), request.fileId)
        fileFolder.deleteRecursively()
        val cachedFile = File(fileFolder, request.fileName)
        return CachedFileDownload(cachedFile) { file ->
            FileProvider.getUriForFile(context, fileProviderAuthority(context), file)
        }
    }
}

class RestoredDownload(val destination: PendingDownload, val writtenBytes: Long)

class DownloadDestinationRestorer @Inject constructor(
    @ApplicationContext private val context: Context,
) {

    fun restore(request: DownloadRequest, reference: String, mimeType: String): RestoredDownload? =
        when (request.purpose) {
            DownloadPurpose.SAVE_TO_DEVICE -> restoreMediaStoreEntry(request, reference, mimeType)
            DownloadPurpose.OPEN_WITH, DownloadPurpose.SHARE -> restoreCachedFile(reference)
        }

    fun discard(request: DownloadRequest, reference: String, mimeType: String) {
        restore(request, reference, mimeType)?.destination?.discard()
    }

    private fun restoreMediaStoreEntry(request: DownloadRequest, reference: String, mimeType: String): RestoredDownload? {
        val destination = MediaStorePendingDownload(context.contentResolver, request.fileName, mimeType, Uri.parse(reference))
        val writtenBytes = destination.writtenBytes() ?: return null
        return RestoredDownload(destination, writtenBytes)
    }

    private fun restoreCachedFile(reference: String): RestoredDownload? {
        val file = File(reference)
        if (!file.isFile) return null
        val destination = CachedFileDownload(file) { cachedFile ->
            FileProvider.getUriForFile(context, fileProviderAuthority(context), cachedFile)
        }
        return RestoredDownload(destination, file.length())
    }
}
