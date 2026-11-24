package com.kuranas.android.feature.transfers.data.upload

sealed interface UploadOutcome {
    data object Completed : UploadOutcome
    data object AlreadyExisted : UploadOutcome
    data object ConnectionLost : UploadOutcome
    data class Rejected(val serverMessage: String?) : UploadOutcome
    data object SourceMissing : UploadOutcome
}

data class UploadRequest(
    val uploadId: String?,
    val targetFolderId: Long?,
    val displayName: String,
    val relativePath: String? = null,
)

fun interface UploadProgressListener {
    suspend fun onProgress(uploadId: String, receivedBytes: Long)
}

fun interface UploadChunkGate {
    suspend fun beforeChunk()
}
