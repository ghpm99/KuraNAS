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
        val transfer = transferDao.findById(transferId) ?: return
        val transition = stateMachine.cancel(transfer) as? TransferTransition.Applied ?: return
        transferDao.update(transition.transfer)
        scheduler.cancelWork(transferId)
        transfer.uploadId?.let { discardServerSession(it) }
        transfer.stagedFilePath?.let { File(it).parentFile?.deleteRecursively() }
    }

    suspend fun retry(transferId: Long) {
        val transfer = transferDao.findById(transferId) ?: return
        val transition = stateMachine.retry(transfer) as? TransferTransition.Applied ?: return
        transferDao.update(transition.transfer)
        scheduler.enqueueUpload(transferId, ExistingWorkPolicy.REPLACE)
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
