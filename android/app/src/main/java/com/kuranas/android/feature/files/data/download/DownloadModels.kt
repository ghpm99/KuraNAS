package com.kuranas.android.feature.files.data.download

import android.net.Uri

enum class DownloadPurpose { SAVE_TO_DEVICE, OPEN_WITH, SHARE }

data class DownloadRequest(
    val fileId: String,
    val fileName: String,
    val sizeBytes: Long,
    val purpose: DownloadPurpose = DownloadPurpose.SAVE_TO_DEVICE,
)

data class DownloadProgress(val receivedBytes: Long, val totalBytes: Long) {
    val fraction: Float?
        get() = if (totalBytes > 0) (receivedBytes.toFloat() / totalBytes).coerceIn(0f, 1f) else null
}

sealed interface DownloadFailure {
    data class Server(val message: String?, val httpCode: Int) : DownloadFailure
    data object ConnectionLost : DownloadFailure
    data object StorageUnavailable : DownloadFailure
}

sealed interface DownloadOutcome {
    data class Completed(val uri: Uri, val mimeType: String) : DownloadOutcome
    data class Failed(val failure: DownloadFailure) : DownloadOutcome
}

sealed interface DownloadState {
    data object Queued : DownloadState
    data class Running(val progress: DownloadProgress) : DownloadState
    data class Completed(val uri: Uri, val mimeType: String) : DownloadState
    data class Failed(val failure: DownloadFailure) : DownloadState
}

data class DownloadEntry(val id: Long, val request: DownloadRequest, val state: DownloadState)
