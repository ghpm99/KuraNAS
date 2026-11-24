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
        val destinations = newDestinationFactory(request, mediaStoreDestinations, cacheDestinations)
        return downloader.download(request, destinations, onProgress = { onProgress(it) })
    }
}

data class PersistedDownloadDestination(val reference: String, val mimeType: String)

class ResumableDownloadExecutor @Inject constructor(
    private val downloader: StreamingDownloader,
    private val mediaStoreDestinations: MediaStoreDownloadDestinations,
    private val cacheDestinations: CacheDownloadDestinations,
    private val restorer: DownloadDestinationRestorer,
) {

    suspend fun execute(
        request: DownloadRequest,
        persistedDestination: PersistedDownloadDestination?,
        onDestinationReady: suspend (reference: String, mimeType: String) -> Unit,
        onProgress: suspend (DownloadProgress) -> Unit,
    ): DownloadOutcome = downloader.download(
        request = request,
        destinations = newDestinationFactory(request, mediaStoreDestinations, cacheDestinations),
        resume = restoreResume(request, persistedDestination),
        shouldKeepPartialForLaterResume = true,
        onDestinationReady = onDestinationReady,
        onProgress = onProgress,
    )

    private fun restoreResume(request: DownloadRequest, persisted: PersistedDownloadDestination?): DownloadResume? {
        persisted ?: return null
        val restored = restorer.restore(request, persisted.reference, persisted.mimeType) ?: return null
        val isAlreadyWhole = request.sizeBytes > 0 && restored.writtenBytes >= request.sizeBytes
        if (isAlreadyWhole) {
            restored.destination.discard()
            return null
        }
        return DownloadResume(restored.destination, restored.writtenBytes, persisted.mimeType)
    }
}

private fun newDestinationFactory(
    request: DownloadRequest,
    mediaStoreDestinations: MediaStoreDownloadDestinations,
    cacheDestinations: CacheDownloadDestinations,
) = PendingDownloadFactory { mimeType ->
    when (request.purpose) {
        DownloadPurpose.SAVE_TO_DEVICE -> mediaStoreDestinations.create(request, mimeType)
        DownloadPurpose.OPEN_WITH, DownloadPurpose.SHARE -> cacheDestinations.create(request)
    }
}
