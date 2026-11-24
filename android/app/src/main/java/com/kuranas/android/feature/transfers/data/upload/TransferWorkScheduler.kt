package com.kuranas.android.feature.transfers.data.upload

import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.workDataOf
import java.util.concurrent.TimeUnit
import javax.inject.Inject

class TransferWorkScheduler @Inject constructor(private val workManager: WorkManager) {

    fun enqueueUpload(transferId: Long, existingWorkPolicy: ExistingWorkPolicy) {
        val request = OneTimeWorkRequestBuilder<UploadTransferWorker>()
            .setInputData(workDataOf(UploadTransferWorker.KEY_TRANSFER_ID to transferId))
            .setConstraints(Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build())
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, INITIAL_BACKOFF_SECONDS, TimeUnit.SECONDS)
            .addTag(TRANSFERS_TAG)
            .build()
        workManager.enqueueUniqueWork(uniqueNameOf(transferId), existingWorkPolicy, request)
    }

    fun cancelWork(transferId: Long) {
        workManager.cancelUniqueWork(uniqueNameOf(transferId))
    }

    private fun uniqueNameOf(transferId: Long) = "transfer-$transferId"

    companion object {
        const val TRANSFERS_TAG = "transfers"
        private const val INITIAL_BACKOFF_SECONDS = 10L
    }
}
