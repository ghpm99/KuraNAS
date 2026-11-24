package com.kuranas.android.feature.transfers.data.upload

import android.content.ContentResolver
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.work.ExistingWorkPolicy
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.data.describeSelectedFile
import com.kuranas.android.feature.transfers.data.TransferClock
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferTransition
import com.kuranas.android.feature.transfers.data.download.PartialDownloadDiscarder
import com.kuranas.android.feature.transfers.data.download.withoutPartialDownload
import com.kuranas.android.feature.transfers.data.notification.TransferBatchNotifier
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.io.IOException
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

data class UploadRequestItem(
    val displayName: String,
    val sizeBytes: Long,
    val sourceUri: String? = null,
    val stagedFilePath: String? = null,
    val relativePath: String? = null,
)

@Singleton
class TransferRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val contentResolver: ContentResolver,
    private val transferDao: TransferDao,
    private val stateMachine: TransferStateMachine,
    private val scheduler: TransferWorkScheduler,
    private val uploadApi: TransferUploadApi,
    private val clock: TransferClock,
    private val partialDownloadDiscarder: PartialDownloadDiscarder,
    private val batchNotifier: TransferBatchNotifier,
) {

    suspend fun enqueueUploads(items: List<UploadRequestItem>, destination: UploadDestination): String {
        val batchId = UUID.randomUUID().toString()
        val now = clock.nowMillis()
        val entities = items.map { item ->
            TransferEntity(
                batchId = batchId,
                direction = TransferDirection.UPLOAD,
                displayName = item.displayName,
                sourceUri = item.sourceUri,
                stagedFilePath = item.stagedFilePath,
                targetFolderId = destination.folderId?.toLongOrNull(),
                targetFolderName = destination.folderName,
                relativePath = item.relativePath,
                totalBytes = item.sizeBytes,
                createdAt = now,
                updatedAt = now,
            )
        }
        val transferIds = transferDao.insertAll(entities)
        transferIds.forEach { scheduler.enqueueUpload(it, ExistingWorkPolicy.KEEP) }
        return batchId
    }

    suspend fun stageSharedUri(uri: Uri): String = withContext(Dispatchers.IO) {
        val selectedFile = describeSelectedFile(context, contentResolver, uri)
        val stagingDirectory = File(File(context.filesDir, OUTBOX_DIRECTORY), UUID.randomUUID().toString())
        stagingDirectory.mkdirs()
        val stagedFile = File(stagingDirectory, File(selectedFile.name).name)
        try {
            val input = contentResolver.openInputStream(uri) ?: throw IOException(uri.toString())
            input.use { stream -> stagedFile.outputStream().use { stream.copyTo(it) } }
        } catch (exception: IOException) {
            stagingDirectory.deleteRecursively()
            throw exception
        } catch (exception: SecurityException) {
            stagingDirectory.deleteRecursively()
            throw exception
        }
        stagedFile.absolutePath
    }

    suspend fun persistPickedUri(uri: Uri): UploadRequestItem = withContext(Dispatchers.IO) {
        try {
            contentResolver.takePersistableUriPermission(uri, Intent.FLAG_GRANT_READ_URI_PERMISSION)
        } catch (exception: SecurityException) {
            return@withContext stagedItemOf(uri)
        }
        val selectedFile = describeSelectedFile(context, contentResolver, uri)
        UploadRequestItem(selectedFile.name, selectedFile.sizeBytes, sourceUri = uri.toString())
    }

    suspend fun cancel(transferId: Long) {
        var transferBeforeCancel: TransferEntity? = null
        val cancelled = transferDao.applyTransition(transferId) { current ->
            transferBeforeCancel = current
            cancelledTransition(current)
        } ?: return
        val previous = transferBeforeCancel ?: return
        scheduler.cancelWork(transferId)
        if (previous.direction == TransferDirection.DOWNLOAD) partialDownloadDiscarder.discard(previous)
        cancelled.uploadId?.let { discardServerSession(it) }
        cancelled.stagedFilePath?.let { File(it).parentFile?.deleteRecursively() }
        batchNotifier.refresh(cancelled.batchId)
    }

    private fun cancelledTransition(current: TransferEntity): TransferTransition =
        when (val transition = stateMachine.cancel(current)) {
            is TransferTransition.Rejected -> transition
            is TransferTransition.Applied ->
                if (current.direction == TransferDirection.DOWNLOAD) {
                    TransferTransition.Applied(transition.transfer.withoutPartialDownload())
                } else {
                    transition
                }
        }

    suspend fun retry(transferId: Long) {
        val queued = transferDao.applyTransition(transferId) { stateMachine.retry(it) } ?: return
        scheduler.enqueue(transferId, queued.direction, ExistingWorkPolicy.REPLACE)
    }

    suspend fun clearFinished() = transferDao.deleteFinished()

    private suspend fun stagedItemOf(uri: Uri): UploadRequestItem {
        val stagedPath = stageSharedUri(uri)
        val stagedFile = File(stagedPath)
        return UploadRequestItem(stagedFile.name, stagedFile.length(), stagedFilePath = stagedPath)
    }

    private suspend fun discardServerSession(uploadId: String) {
        try {
            uploadApi.cancelUpload(uploadId)
        } catch (exception: IOException) {
            return
        }
    }

    private companion object {
        const val OUTBOX_DIRECTORY = "outbox"
    }
}
