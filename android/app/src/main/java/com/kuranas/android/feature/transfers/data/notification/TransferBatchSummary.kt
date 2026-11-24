package com.kuranas.android.feature.transfers.data.notification

import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus

data class TransferBatchSummary(
    val direction: TransferDirection,
    val totalCount: Int,
    val finishedCount: Int,
    val completedCount: Int,
    val alreadyExistedCount: Int,
    val failedCount: Int,
    val percent: Int,
    val isFinished: Boolean,
    val isWaitingForConnection: Boolean,
    val singleSavedDownload: TransferEntity?,
    val hasSavedDownload: Boolean,
)

fun summarizeBatch(transfers: List<TransferEntity>): TransferBatchSummary? {
    val firstTransfer = transfers.firstOrNull() ?: return null
    val countedTransfers = transfers.filter { it.status != TransferStatus.CANCELLED }
    val finishedCount = countedTransfers.count { !it.status.isActive }
    val savedDownloads = if (firstTransfer.direction == TransferDirection.DOWNLOAD) {
        countedTransfers.filter { it.isSavedToDevice() }
    } else {
        emptyList()
    }
    return TransferBatchSummary(
        direction = firstTransfer.direction,
        totalCount = countedTransfers.size,
        finishedCount = finishedCount,
        completedCount = countedTransfers.count { it.status == TransferStatus.COMPLETED },
        alreadyExistedCount = countedTransfers.count { it.status == TransferStatus.ALREADY_EXISTED },
        failedCount = countedTransfers.count { it.status == TransferStatus.FAILED },
        percent = batchPercent(countedTransfers, finishedCount),
        isFinished = transfers.none { it.status.isActive },
        isWaitingForConnection = isWaitingForConnection(transfers),
        singleSavedDownload = savedDownloads.singleOrNull()
            ?.takeIf { countedTransfers.size == 1 && it.status == TransferStatus.COMPLETED && it.destinationUri != null },
        hasSavedDownload = savedDownloads.isNotEmpty(),
    )
}

private fun TransferEntity.isSavedToDevice(): Boolean =
    downloadPurpose == null || downloadPurpose == DownloadPurpose.SAVE_TO_DEVICE.name

private fun isWaitingForConnection(transfers: List<TransferEntity>): Boolean {
    val hasRunning = transfers.any { it.status == TransferStatus.RUNNING }
    return !hasRunning && transfers.any { it.status == TransferStatus.WAITING_FOR_NETWORK }
}

private fun batchPercent(countedTransfers: List<TransferEntity>, finishedCount: Int): Int {
    if (countedTransfers.isEmpty()) return FULL_PERCENT
    val areAllTotalsKnown = countedTransfers.all { it.totalBytes > 0 }
    if (!areAllTotalsKnown) return finishedCount * FULL_PERCENT / countedTransfers.size
    val totalBytes = countedTransfers.sumOf { it.totalBytes }
    val doneBytes = countedTransfers.sumOf { bytesDone(it) }
    return (doneBytes * FULL_PERCENT / totalBytes).toInt()
}

private fun bytesDone(transfer: TransferEntity): Long = when (transfer.status) {
    TransferStatus.COMPLETED, TransferStatus.ALREADY_EXISTED -> transfer.totalBytes
    else -> transfer.transferredBytes.coerceIn(0, transfer.totalBytes)
}

private const val FULL_PERCENT = 100
