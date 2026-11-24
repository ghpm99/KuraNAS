package com.kuranas.android.feature.transfers.data.download

import android.net.Uri
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.TransferTransition
import io.mockk.every
import io.mockk.mockk
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class DownloadOutcomeResolverTest {

    private val resolver = DownloadOutcomeResolver(
        TransferStateMachine { 9_000L },
        DownloadFailureMessages(storageUnavailable = "no storage", serverRejectedWithoutDetail = { "http $it" }),
    )

    private val runningTransfer = TransferEntity(
        id = 1,
        batchId = "batch",
        direction = TransferDirection.DOWNLOAD,
        displayName = "movie.mkv",
        targetFolderName = "",
        remoteFileId = 7,
        totalBytes = 100,
        transferredBytes = 40,
        destinationUri = "content://media/pending/1",
        mimeType = "video/x-matroska",
        status = TransferStatus.RUNNING,
        createdAt = 1,
        updatedAt = 1,
    )

    private fun resolved(outcome: DownloadOutcome): Pair<TransferEntity, Boolean> {
        val resolution = resolver.resolve(runningTransfer, outcome)
        return (resolution.transition as TransferTransition.Applied).transfer to resolution.shouldRetryLater
    }

    @Test
    fun completedStoresTheFinalUriAndMimeType() {
        val finalUri = mockk<Uri>()
        every { finalUri.toString() } returns "content://media/final/1"

        val (transfer, shouldRetry) = resolved(DownloadOutcome.Completed(finalUri, "video/mp4"))

        assertEquals(TransferStatus.COMPLETED, transfer.status)
        assertEquals("content://media/final/1", transfer.destinationUri)
        assertEquals("video/mp4", transfer.mimeType)
        assertEquals(100, transfer.transferredBytes)
        assertFalse(shouldRetry)
    }

    @Test
    fun connectionLostWaitsForNetworkKeepingThePartialDestination() {
        val (transfer, shouldRetry) = resolved(DownloadOutcome.Failed(DownloadFailure.ConnectionLost))

        assertEquals(TransferStatus.WAITING_FOR_NETWORK, transfer.status)
        assertEquals("content://media/pending/1", transfer.destinationUri)
        assertTrue(shouldRetry)
    }

    @Test
    fun serverFailureShowsTheServerMessageVerbatimAndForgetsThePartial() {
        val (transfer, shouldRetry) = resolved(DownloadOutcome.Failed(DownloadFailure.Server("Arquivo não encontrado", 404)))

        assertEquals(TransferStatus.FAILED, transfer.status)
        assertEquals("Arquivo não encontrado", transfer.failureMessage)
        assertNull(transfer.destinationUri)
        assertEquals(0, transfer.transferredBytes)
        assertFalse(shouldRetry)
    }

    @Test
    fun serverFailureWithoutMessageUsesTheHttpCodeText() {
        val (transfer, _) = resolved(DownloadOutcome.Failed(DownloadFailure.Server(null, 502)))

        assertEquals("http 502", transfer.failureMessage)
    }

    @Test
    fun storageUnavailableFailsWithTheStorageMessage() {
        val (transfer, shouldRetry) = resolved(DownloadOutcome.Failed(DownloadFailure.StorageUnavailable))

        assertEquals(TransferStatus.FAILED, transfer.status)
        assertEquals("no storage", transfer.failureMessage)
        assertFalse(shouldRetry)
    }

    @Test
    fun anOutcomeForAnAlreadyCancelledTransferIsRejected() {
        val cancelled = runningTransfer.copy(status = TransferStatus.CANCELLED)

        val resolution = resolver.resolve(cancelled, DownloadOutcome.Failed(DownloadFailure.ConnectionLost))

        assertTrue(resolution.transition is TransferTransition.Rejected)
    }
}
