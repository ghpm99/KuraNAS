package com.kuranas.android.feature.transfers.data.notification

import com.kuranas.android.feature.transfers.data.TransferDirection

enum class BatchResultPartKind { TRANSFERRED, ALREADY_EXISTED, FAILED }

data class BatchResultPart(val kind: BatchResultPartKind, val count: Int)

sealed interface TransferBatchNotificationPlan {
    data object Nothing : TransferBatchNotificationPlan
    data class Progress(val summary: TransferBatchSummary) : TransferBatchNotificationPlan
    data class Result(val summary: TransferBatchSummary) : TransferBatchNotificationPlan
}

fun planBatchNotification(summary: TransferBatchSummary?): TransferBatchNotificationPlan {
    if (summary == null) return TransferBatchNotificationPlan.Nothing
    if (!summary.isFinished) return TransferBatchNotificationPlan.Progress(summary)
    if (summary.totalCount == 0) return TransferBatchNotificationPlan.Nothing
    val isViewerOnlyDownloadBatch = summary.direction == TransferDirection.DOWNLOAD &&
        !summary.hasSavedDownload
    if (isViewerOnlyDownloadBatch) return TransferBatchNotificationPlan.Nothing
    return TransferBatchNotificationPlan.Result(summary)
}

fun batchResultParts(summary: TransferBatchSummary): List<BatchResultPart> = listOf(
    BatchResultPart(BatchResultPartKind.TRANSFERRED, summary.completedCount),
    BatchResultPart(BatchResultPartKind.ALREADY_EXISTED, summary.alreadyExistedCount),
    BatchResultPart(BatchResultPartKind.FAILED, summary.failedCount),
).filter { it.count > 0 }
