package com.kuranas.android.feature.transfers.data.download

import android.net.Uri
import com.kuranas.android.feature.files.data.download.DownloadEntry
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.feature.files.data.download.DownloadState
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus
import io.mockk.mockk
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class DownloadTransferMappingTest {

    private val parsedUri = mockk<Uri>()
    private val parseUri: (String) -> Uri = { parsedUri }

    private val queuedDownload = TransferEntity(
        id = 5,
        batchId = "batch",
        direction = TransferDirection.DOWNLOAD,
        displayName = "relatório.pdf",
        targetFolderName = "",
        remoteFileId = 42,
        downloadPurpose = DownloadPurpose.SAVE_TO_DEVICE.name,
        totalBytes = 1_000,
        createdAt = 100,
        updatedAt = 100,
    )

    private fun entryOf(transfer: TransferEntity, visibleSince: Long = 0) =
        transfer.toDownloadEntryOrNull(visibleSince, parseUri)

    @Test
    fun queuedAndWaitingForNetworkAreQueued() {
        assertEquals(DownloadState.Queued, entryOf(queuedDownload)?.state)
        assertEquals(DownloadState.Queued, entryOf(queuedDownload.copy(status = TransferStatus.WAITING_FOR_NETWORK))?.state)
    }

    @Test
    fun runningCarriesTheProgress() {
        val entry = entryOf(queuedDownload.copy(status = TransferStatus.RUNNING, transferredBytes = 250))

        assertEquals(
            DownloadEntry(5, DownloadRequest("42", "relatório.pdf", 1_000), DownloadState.Running(DownloadProgress(250, 1_000))),
            entry,
        )
    }

    @Test
    fun completedCarriesTheUriAndMimeType() {
        val entry = entryOf(
            queuedDownload.copy(status = TransferStatus.COMPLETED, destinationUri = "content://x", mimeType = "application/pdf"),
        )

        assertEquals(DownloadState.Completed(parsedUri, "application/pdf"), entry?.state)
    }

    @Test
    fun completedWithoutDestinationIsFailedStorage() {
        assertEquals(
            DownloadState.Failed(DownloadFailure.StorageUnavailable),
            entryOf(queuedDownload.copy(status = TransferStatus.COMPLETED))?.state,
        )
    }

    @Test
    fun failedShowsTheStoredMessageVerbatim() {
        val entry = entryOf(queuedDownload.copy(status = TransferStatus.FAILED, failureMessage = "Sem espaço"))

        assertEquals(DownloadState.Failed(DownloadFailure.Server("Sem espaço", 0)), entry?.state)
    }

    @Test
    fun cancelledUploadsAndUnparsableRowsAreHidden() {
        assertNull(entryOf(queuedDownload.copy(status = TransferStatus.CANCELLED)))
        assertNull(entryOf(queuedDownload.copy(direction = TransferDirection.UPLOAD)))
        assertNull(entryOf(queuedDownload.copy(remoteFileId = null)))
    }

    @Test
    fun finishedRowsFromBeforeTheSessionAreHiddenButActiveOnesStay() {
        val oldCompleted = queuedDownload.copy(status = TransferStatus.COMPLETED, destinationUri = "u", mimeType = "m", updatedAt = 50)

        assertNull(entryOf(oldCompleted, visibleSince = 100))
        assertEquals(DownloadState.Queued, entryOf(queuedDownload.copy(updatedAt = 50), visibleSince = 100)?.state)
    }

    @Test
    fun purposeDefaultsToSaveWhenUnknown() {
        assertEquals(DownloadPurpose.SAVE_TO_DEVICE, queuedDownload.copy(downloadPurpose = "other").toDownloadRequest().purpose)
        assertEquals(DownloadPurpose.SHARE, queuedDownload.copy(downloadPurpose = "SHARE").toDownloadRequest().purpose)
    }
}
