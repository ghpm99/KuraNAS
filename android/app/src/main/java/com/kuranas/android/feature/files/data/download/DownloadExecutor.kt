package com.kuranas.android.feature.files.data.download

import javax.inject.Inject

interface DownloadExecutor {
    suspend fun execute(request: DownloadRequest, onProgress: (DownloadProgress) -> Unit): DownloadOutcome
}

class StreamingDownloadExecutor @Inject constructor(
    private val downloader: StreamingDownloader,
    private val mediaStoreDestinations: MediaStoreDownloadDestinations,
    private val cacheDestinations: CacheDownloadDestinations,
) : DownloadExecutor {

    override suspend fun execute(
        request: DownloadRequest,
        onProgress: (DownloadProgress) -> Unit,
    ): DownloadOutcome {
        val destinations = PendingDownloadFactory { mimeType ->
            when (request.purpose) {
                DownloadPurpose.SAVE_TO_DEVICE -> mediaStoreDestinations.create(request, mimeType)
                DownloadPurpose.OPEN_WITH, DownloadPurpose.SHARE -> cacheDestinations.create(request)
            }
        }
        return downloader.download(request, destinations, onProgress)
    }
}
