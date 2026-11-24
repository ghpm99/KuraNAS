package com.kuranas.android.feature.transfers.ui

import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus

data class TransferItem(
    val id: Long,
    val name: String,
    val direction: TransferDirection,
    val destinationFolderName: String?,
    val progressFraction: Float?,
    val status: TransferStatus,
    val failureMessage: String?,
) {
    val canCancel: Boolean get() = status.isActive
    val canRetry: Boolean get() = status == TransferStatus.FAILED || status == TransferStatus.CANCELLED
}

data class TransfersUiState(
    val inProgress: List<TransferItem> = emptyList(),
    val completed: List<TransferItem> = emptyList(),
    val failed: List<TransferItem> = emptyList(),
) {
    val isEmpty: Boolean get() = inProgress.isEmpty() && completed.isEmpty() && failed.isEmpty()
    val hasFinishedItems: Boolean get() = completed.isNotEmpty()
}

fun TransferEntity.toTransferItem(): TransferItem = TransferItem(
    id = id,
    name = displayName,
    direction = direction,
    destinationFolderName = destinationLabel().takeIf { direction == TransferDirection.UPLOAD },
    progressFraction = progressFractionOf(transferredBytes, totalBytes),
    status = status,
    failureMessage = failureMessage,
)

private fun TransferEntity.destinationLabel(): String =
    relativePath?.takeIf { it.isNotBlank() }?.let { "$targetFolderName/$it" } ?: targetFolderName

fun groupTransfers(transfers: List<TransferEntity>): TransfersUiState {
    val items = transfers.map { it.toTransferItem() }
    return TransfersUiState(
        inProgress = items.filter { it.status.isActive },
        completed = items.filter { it.status in COMPLETED_STATUSES },
        failed = items.filter { it.status == TransferStatus.FAILED },
    )
}

private val COMPLETED_STATUSES = setOf(TransferStatus.COMPLETED, TransferStatus.ALREADY_EXISTED, TransferStatus.CANCELLED)

private fun progressFractionOf(transferredBytes: Long, totalBytes: Long): Float? {
    if (totalBytes <= 0) return null
    return (transferredBytes.toFloat() / totalBytes).coerceIn(0f, 1f)
}
