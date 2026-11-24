package com.kuranas.android.feature.transfers.data.upload

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.kuranas.android.R
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.TransferTransition
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
) : CoroutineWorker(appContext, workerParameters) {

    override suspend fun doWork(): Result {
        val transferId = inputData.getLong(KEY_TRANSFER_ID, NO_TRANSFER_ID)
        val storedTransfer = transferDao.findById(transferId) ?: return Result.success()
        if (!storedTransfer.status.isActive) return Result.success()
        val runningTransfer = markRunning(storedTransfer) ?: return Result.success()

        val outcome = try {
            uploadWithPersistedProgress(runningTransfer)
        } catch (cancelled: TransferCancelledException) {
            return Result.success()
        }
        return finish(transferId, outcome)
    }

    private suspend fun markRunning(transfer: TransferEntity): TransferEntity? {
        if (transfer.status == TransferStatus.RUNNING) return transfer
        val transition = stateMachine.start(transfer) as? TransferTransition.Applied ?: return null
        transferDao.update(transition.transfer)
        return transition.transfer
    }

    private suspend fun uploadWithPersistedProgress(transfer: TransferEntity): UploadOutcome {
        val source = try {
            sourceFactory.open(transfer)
        } catch (exception: UploadSourceUnavailableException) {
            return UploadOutcome.SourceMissing
        }
        val request = UploadRequest(transfer.uploadId, transfer.targetFolderId, transfer.displayName)
        return uploader.upload(
            source = source,
            request = request,
            progressListener = { uploadId, receivedBytes -> persistProgress(transfer.id, uploadId, receivedBytes) },
            chunkGate = { requireStillActive(transfer.id) },
        )
    }

    private suspend fun persistProgress(transferId: Long, uploadId: String, receivedBytes: Long) {
        val current = requireStillActive(transferId)
        val progressed = stateMachine.progress(current, receivedBytes) as? TransferTransition.Applied
            ?: throw TransferCancelledException()
        transferDao.update(progressed.transfer.copy(uploadId = uploadId))
    }

    private suspend fun requireStillActive(transferId: Long): TransferEntity {
        val current = transferDao.findById(transferId)
        if (current == null || !current.status.isActive) throw TransferCancelledException()
        return current
    }

    private suspend fun finish(transferId: Long, outcome: UploadOutcome): Result {
        val current = transferDao.findById(transferId) ?: return Result.success()
        if (!current.status.isActive) return Result.success()
        val resolver = UploadOutcomeResolver(
            stateMachine,
            UploadFailureMessages(
                sourceMissing = applicationContext.getString(R.string.transfer_source_missing),
                rejectedWithoutDetail = applicationContext.getString(R.string.transfer_rejected_without_detail),
            ),
        )
        val resolution = resolver.resolve(current, outcome)
        val applied = resolution.transition as? TransferTransition.Applied ?: return Result.success()
        transferDao.update(applied.transfer)
        if (resolution.shouldRetryLater) return Result.retry()
        if (applied.transfer.status != TransferStatus.FAILED) deleteStagedFile(applied.transfer)
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
