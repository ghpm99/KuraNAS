package com.kuranas.android.feature.transfers.data.upload

import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequest
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.workDataOf
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.download.DownloadTransferWorker
import java.util.concurrent.TimeUnit
import javax.inject.Inject

class TransferWorkScheduler @Inject constructor(private val workManager: WorkManager) {

    fun enqueueUpload(transferId: Long, existingWorkPolicy: ExistingWorkPolicy, requiresUnmeteredNetwork: Boolean = false) =
        enqueue(
            OneTimeWorkRequestBuilder<UploadTransferWorker>(),
            UploadTransferWorker.KEY_TRANSFER_ID,
            transferId,
            existingWorkPolicy,
            requiredNetworkTypeOf(requiresUnmeteredNetwork),
        )

    fun enqueueDownload(transferId: Long, existingWorkPolicy: ExistingWorkPolicy) =
        enqueue(OneTimeWorkRequestBuilder<DownloadTransferWorker>(), DownloadTransferWorker.KEY_TRANSFER_ID, transferId, existingWorkPolicy, NetworkType.CONNECTED)

    fun enqueue(
        transferId: Long,
        direction: TransferDirection,
        existingWorkPolicy: ExistingWorkPolicy,
        requiresUnmeteredNetwork: Boolean = false,
    ) =
        when (direction) {
            TransferDirection.UPLOAD -> enqueueUpload(transferId, existingWorkPolicy, requiresUnmeteredNetwork)
            TransferDirection.DOWNLOAD -> enqueueDownload(transferId, existingWorkPolicy)
        }

    private fun enqueue(
        requestBuilder: OneTimeWorkRequest.Builder,
        transferIdKey: String,
        transferId: Long,
        existingWorkPolicy: ExistingWorkPolicy,
        networkType: NetworkType,
    ) {
        val request = requestBuilder
            .setInputData(workDataOf(transferIdKey to transferId))
            .setConstraints(Constraints.Builder().setRequiredNetworkType(networkType).build())
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, INITIAL_BACKOFF_SECONDS, TimeUnit.SECONDS)
            .addTag(TRANSFERS_TAG)
            .build()
        workManager.enqueueUniqueWork(uniqueNameOf(transferId), existingWorkPolicy, request)
    }

    fun cancelWork(transferId: Long) {
        workManager.cancelUniqueWork(uniqueNameOf(transferId))
    }

    private fun requiredNetworkTypeOf(requiresUnmeteredNetwork: Boolean) =
        if (requiresUnmeteredNetwork) NetworkType.UNMETERED else NetworkType.CONNECTED

    private fun uniqueNameOf(transferId: Long) = "transfer-$transferId"

    companion object {
        const val TRANSFERS_TAG = "transfers"
        private const val INITIAL_BACKOFF_SECONDS = 10L
    }
}
