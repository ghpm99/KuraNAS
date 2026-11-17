package com.kuranas.android.feature.files.data.download

import android.content.Context
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
