package com.kuranas.android.feature.transfers.data.download

import android.net.Uri
import com.kuranas.android.feature.files.data.download.DownloadEntry
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.feature.files.data.download.DownloadState
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus

fun TransferEntity.toDownloadRequest(): DownloadRequest = DownloadRequest(
    fileId = remoteFileId.toString(),
    fileName = displayName,
    sizeBytes = totalBytes,
    purpose = downloadPurpose?.let { purposeName -> DownloadPurpose.entries.firstOrNull { it.name == purposeName } }
        ?: DownloadPurpose.SAVE_TO_DEVICE,
)

/**
 * Visão de uma linha da fila para a interface de downloads. Linhas já terminadas antes de [visibleSinceMillis]
 * ficam fora, para que o resultado de uma sessão anterior não reapareça como se fosse do pedido atual.
 */
fun TransferEntity.toDownloadEntryOrNull(visibleSinceMillis: Long, parseUri: (String) -> Uri = Uri::parse): DownloadEntry? {
    if (direction != TransferDirection.DOWNLOAD || remoteFileId == null) return null
    if (!status.isActive && updatedAt < visibleSinceMillis) return null
    val state = toDownloadState(parseUri) ?: return null
    return DownloadEntry(id, toDownloadRequest(), state)
}

private fun TransferEntity.toDownloadState(parseUri: (String) -> Uri): DownloadState? =
    when (status) {
        TransferStatus.QUEUED, TransferStatus.WAITING_FOR_NETWORK -> DownloadState.Queued
        TransferStatus.RUNNING -> DownloadState.Running(DownloadProgress(transferredBytes, totalBytes))
        TransferStatus.COMPLETED -> completedState(parseUri)
        TransferStatus.FAILED -> DownloadState.Failed(DownloadFailure.Server(failureMessage, httpCode = 0))
        TransferStatus.ALREADY_EXISTED, TransferStatus.CANCELLED -> null
    }

private fun TransferEntity.completedState(parseUri: (String) -> Uri): DownloadState {
    val finishedUri = destinationUri ?: return DownloadState.Failed(DownloadFailure.StorageUnavailable)
    val finishedMimeType = mimeType ?: return DownloadState.Failed(DownloadFailure.StorageUnavailable)
    return DownloadState.Completed(parseUri(finishedUri), finishedMimeType)
}
