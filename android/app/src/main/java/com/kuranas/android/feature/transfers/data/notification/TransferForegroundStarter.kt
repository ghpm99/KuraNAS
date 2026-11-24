package com.kuranas.android.feature.transfers.data.notification

import androidx.work.CoroutineWorker

suspend fun CoroutineWorker.enterForegroundIfAllowed(notifier: TransferBatchNotifier, batchId: String) {
    try {
        setForeground(notifier.progressForegroundInfo(batchId))
    } catch (exception: IllegalStateException) {
        return
    }
}
