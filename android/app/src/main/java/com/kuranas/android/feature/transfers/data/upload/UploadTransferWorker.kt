package com.kuranas.android.feature.transfers.data.upload

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.ForegroundInfo
import androidx.work.WorkerParameters
import com.kuranas.android.R
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.TransferTransition
import com.kuranas.android.feature.transfers.data.notification.TransferBatchNotifier
import com.kuranas.android.feature.transfers.data.notification.enterForegroundIfAllowed
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import java.io.File

@HiltWorker
class UploadTransferWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParameters: WorkerParameters,
    private val transferDao: TransferDao,
    private val stateMachine: TransferStateMachine,
    private val uploader: ChunkedUploader,
    private val sourceFactory: UploadSourceFactory,
    private val batchNotifier: TransferBatchNotifier,
) : CoroutineWorker(appContext, workerParameters) {

    override suspend fun doWork(): Result {
        val transferId = inputData.getLong(KEY_TRANSFER_ID, NO_TRANSFER_ID)
        val storedTransfer = transferDao.findById(transferId) ?: return Result.success()
        if (!storedTransfer.status.isActive) return Result.success()
        val runningTransfer = markRunning(storedTransfer) ?: return Result.success()
        enterForegroundIfAllowed(batchNotifier, runningTransfer.batchId)

        val outcome = try {
            uploadWithPersistedProgress(runningTransfer)
        } catch (cancelled: TransferCancelledException) {
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

    private suspend fun uploadWithPersistedProgress(transfer: TransferEntity): UploadOutcome {
        val source = try {
            sourceFactory.open(transfer)
        } catch (exception: UploadSourceUnavailableException) {
            return UploadOutcome.SourceMissing
        }
        val request = UploadRequest(transfer.uploadId, transfer.targetFolderId, transfer.displayName, transfer.relativePath)
        return uploader.upload(
            source = source,
            request = request,
            progressListener = { uploadId, receivedBytes ->
                persistProgress(transfer.id, uploadId, receivedBytes)
                batchNotifier.refresh(transfer.batchId)
            },
            chunkGate = { requireStillActive(transfer.id) },
        )
    }

    private suspend fun persistProgress(transferId: Long, uploadId: String, receivedBytes: Long) {
        transferDao.applyTransition(transferId) { current ->
            when (val progressed = stateMachine.progress(current, receivedBytes)) {
                is TransferTransition.Applied -> TransferTransition.Applied(progressed.transfer.copy(uploadId = uploadId))
                is TransferTransition.Rejected -> progressed
            }
        } ?: throw TransferCancelledException()
    }

    private suspend fun requireStillActive(transferId: Long): TransferEntity {
        val current = transferDao.findById(transferId)
        if (current == null || !current.status.isActive) throw TransferCancelledException()
        return current
    }

    private suspend fun finish(transferId: Long, outcome: UploadOutcome): Result {
        val resolver = UploadOutcomeResolver(
            stateMachine,
            UploadFailureMessages(
                sourceMissing = applicationContext.getString(R.string.transfer_source_missing),
                rejectedWithoutDetail = applicationContext.getString(R.string.transfer_rejected_without_detail),
            ),
        )
        var shouldRetryLater = false
        val finished = transferDao.applyTransition(transferId) { current ->
            val resolution = resolver.resolve(current, outcome)
            shouldRetryLater = resolution.shouldRetryLater
            resolution.transition
        } ?: return Result.success()
        if (finished.status != TransferStatus.FAILED && !shouldRetryLater) deleteStagedFile(finished)
        batchNotifier.refresh(finished.batchId)
        if (shouldRetryLater) return Result.retry()
        return Result.success()
    }

    private fun deleteStagedFile(transfer: TransferEntity) {
        val stagedFilePath = transfer.stagedFilePath ?: return
        File(stagedFilePath).parentFile?.deleteRecursively()
    }

    companion object {
        const val KEY_TRANSFER_ID = "transfer_id"
        private const val NO_TRANSFER_ID = -1L
    }
}
