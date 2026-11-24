package com.kuranas.android.feature.transfers.data.download

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.ForegroundInfo
import androidx.work.WorkerParameters
import com.kuranas.android.R
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.PersistedDownloadDestination
import com.kuranas.android.feature.files.data.download.ResumableDownloadExecutor
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.TransferTransition
import com.kuranas.android.feature.transfers.data.notification.TransferBatchNotifier
import com.kuranas.android.feature.transfers.data.notification.enterForegroundIfAllowed
import com.kuranas.android.feature.transfers.data.upload.TransferCancelledException
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject

@HiltWorker
class DownloadTransferWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParameters: WorkerParameters,
    private val transferDao: TransferDao,
    private val stateMachine: TransferStateMachine,
    private val executor: ResumableDownloadExecutor,
    private val batchNotifier: TransferBatchNotifier,
    private val partialDiscarder: PartialDownloadDiscarder,
) : CoroutineWorker(appContext, workerParameters) {

    override suspend fun doWork(): Result {
        val transferId = inputData.getLong(KEY_TRANSFER_ID, NO_TRANSFER_ID)
        val storedTransfer = transferDao.findById(transferId) ?: return Result.success()
        if (storedTransfer.direction != TransferDirection.DOWNLOAD || !storedTransfer.status.isActive) return Result.success()
        val runningTransfer = markRunning(storedTransfer) ?: return Result.success()
        enterForegroundIfAllowed(batchNotifier, runningTransfer.batchId)

        val outcome = try {
            downloadWithPersistedProgress(runningTransfer)
        } catch (cancelled: TransferCancelledException) {
            partialDiscarder.discard(transferDao.findById(transferId) ?: runningTransfer)
            batchNotifier.refresh(runningTransfer.batchId)
            return Result.success()
        }
        return finish(transferId, outcome)
    }

    override suspend fun getForegroundInfo(): ForegroundInfo {
        val transferId = inputData.getLong(KEY_TRANSFER_ID, NO_TRANSFER_ID)
        val batchId = transferDao.findById(transferId)?.batchId.orEmpty()
        return batchNotifier.progressForegroundInfo(batchId)
    }

    private suspend fun markRunning(transfer: TransferEntity): TransferEntity? {
        if (transfer.status == TransferStatus.RUNNING) return transfer
        return transferDao.applyTransition(transfer.id) { stateMachine.start(it) }
    }

    private suspend fun downloadWithPersistedProgress(transfer: TransferEntity): DownloadOutcome {
        val request = transfer.toDownloadRequest()
        val persistedDestination = persistedDestinationOf(transfer)
        val throttle = DownloadProgressThrottle()
        return executor.execute(
            request = request,
            persistedDestination = persistedDestination,
            onDestinationReady = { reference, mimeType -> persistDestination(transfer, reference, mimeType) },
            onProgress = { progress ->
                if (throttle.shouldPersist(progress)) {
                    persistProgress(transfer.id, progress)
                    batchNotifier.refresh(transfer.batchId)
                }
            },
        )
    }

    private fun persistedDestinationOf(transfer: TransferEntity): PersistedDownloadDestination? {
        val reference = transfer.destinationUri ?: return null
        val mimeType = transfer.mimeType ?: return null
        return PersistedDownloadDestination(reference, mimeType)
    }

    private suspend fun persistDestination(transfer: TransferEntity, reference: String, mimeType: String) {
        val transferId = transfer.id
        val persisted = transferDao.applyTransition(transferId) { current ->
            if (!current.status.isActive) {
                TransferTransition.Rejected(current, current.status)
            } else {
                TransferTransition.Applied(current.copy(destinationUri = reference, mimeType = mimeType))
            }
        }
        if (persisted == null) {
            partialDiscarder.discard(transfer.copy(destinationUri = reference, mimeType = mimeType))
            throw TransferCancelledException()
        }
    }

    private suspend fun persistProgress(transferId: Long, progress: DownloadProgress) {
        transferDao.applyTransition(transferId) { stateMachine.progress(it, progress.receivedBytes) }
            ?: throw TransferCancelledException()
    }

    private suspend fun finish(transferId: Long, outcome: DownloadOutcome): Result {
        val resolver = newResolver()
        var shouldRetryLater = false
        val finished = transferDao.applyTransition(transferId) { current ->
            val resolution = resolver.resolve(current, outcome)
            shouldRetryLater = resolution.shouldRetryLater
            resolution.transition
        } ?: return Result.success()
        batchNotifier.refresh(finished.batchId)
        if (shouldRetryLater) return Result.retry()
        return Result.success()
    }

    private fun newResolver() = DownloadOutcomeResolver(
        stateMachine,
        DownloadFailureMessages(
            storageUnavailable = applicationContext.getString(R.string.download_failed_storage),
            serverRejectedWithoutDetail = { httpCode -> applicationContext.getString(R.string.download_failed_http, httpCode) },
        ),
    )

    companion object {
        const val KEY_TRANSFER_ID = "transfer_id"
        private const val NO_TRANSFER_ID = -1L
    }
}
