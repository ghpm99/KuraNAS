package com.kuranas.android.feature.transfers.data.download

import androidx.work.ExistingWorkPolicy
import com.kuranas.android.feature.files.data.download.DownloadEntry
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.feature.transfers.data.TransferClock
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.upload.TransferWorkScheduler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import java.util.UUID

class PersistentDownloadQueue(
    scope: CoroutineScope,
    private val transferDao: TransferDao,
    private val scheduler: TransferWorkScheduler,
    private val clock: TransferClock,
) : DownloadQueue {

    private val sessionStartMillis = clock.nowMillis()

    override val entries: StateFlow<List<DownloadEntry>> = transferDao.observeAll()
        .map { transfers -> transfers.mapNotNull { it.toDownloadEntryOrNull(sessionStartMillis) }.sortedBy { it.id } }
        .stateIn(scope, SharingStarted.Eagerly, emptyList())

    override suspend fun enqueue(requests: List<DownloadRequest>): List<Long> {
        val batchId = UUID.randomUUID().toString()
        val now = clock.nowMillis()
        val transferIds = transferDao.insertAll(requests.map { newDownloadTransfer(batchId, it, now) })
        transferIds.forEach { scheduler.enqueueDownload(it, ExistingWorkPolicy.KEEP) }
        return transferIds
    }

    private fun newDownloadTransfer(batchId: String, request: DownloadRequest, now: Long) = TransferEntity(
        batchId = batchId,
        direction = TransferDirection.DOWNLOAD,
        displayName = request.fileName,
        targetFolderName = "",
        remoteFileId = request.fileId.toLongOrNull(),
        downloadPurpose = request.purpose.name,
        totalBytes = request.sizeBytes,
        createdAt = now,
        updatedAt = now,
    )
}
