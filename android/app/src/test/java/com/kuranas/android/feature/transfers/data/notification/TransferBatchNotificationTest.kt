package com.kuranas.android.feature.transfers.data.notification

import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TransferBatchNotificationTest {

    private fun transfer(
        status: TransferStatus,
        direction: TransferDirection = TransferDirection.UPLOAD,
        totalBytes: Long = 100,
        transferredBytes: Long = 0,
        purpose: String? = null,
        destinationUri: String? = null,
    ) = TransferEntity(
        batchId = "batch",
        direction = direction,
        displayName = "file",
        targetFolderName = "folder",
        downloadPurpose = purpose,
        destinationUri = destinationUri,
        totalBytes = totalBytes,
        transferredBytes = transferredBytes,
        status = status,
        createdAt = 0,
        updatedAt = 0,
    )

    @Test
    fun emptyBatchHasNoSummary() {
        assertNull(summarizeBatch(emptyList()))
    }

    @Test
    fun percentIsWeightedByBytesWhenAllTotalsAreKnown() {
        val summary = summarizeBatch(
            listOf(
                transfer(TransferStatus.COMPLETED, totalBytes = 300),
                transfer(TransferStatus.RUNNING, totalBytes = 100, transferredBytes = 50),
            ),
        )!!
        assertEquals(87, summary.percent)
        assertEquals(1, summary.finishedCount)
        assertEquals(2, summary.totalCount)
        assertFalse(summary.isFinished)
    }

    @Test
    fun percentFallsBackToCountWhenATotalIsUnknown() {
        val summary = summarizeBatch(
            listOf(
                transfer(TransferStatus.COMPLETED, totalBytes = -1),
                transfer(TransferStatus.RUNNING, totalBytes = 100, transferredBytes = 99),
            ),
        )!!
        assertEquals(50, summary.percent)
    }

    @Test
    fun cancelledItemsAreLeftOutOfTheCounts() {
        val summary = summarizeBatch(
            listOf(transfer(TransferStatus.CANCELLED), transfer(TransferStatus.RUNNING)),
        )!!
        assertEquals(1, summary.totalCount)
        assertEquals(0, summary.finishedCount)
    }

    @Test
    fun batchIsWaitingOnlyWhenNothingIsRunning() {
        val waiting = summarizeBatch(listOf(transfer(TransferStatus.WAITING_FOR_NETWORK), transfer(TransferStatus.QUEUED)))!!
        val running = summarizeBatch(listOf(transfer(TransferStatus.WAITING_FOR_NETWORK), transfer(TransferStatus.RUNNING)))!!
        assertTrue(waiting.isWaitingForConnection)
        assertFalse(running.isWaitingForConnection)
    }

    @Test
    fun activeBatchPlansProgress() {
        val plan = planBatchNotification(summarizeBatch(listOf(transfer(TransferStatus.RUNNING))))
        assertTrue(plan is TransferBatchNotificationPlan.Progress)
    }

    @Test
    fun finishedUploadBatchPlansResult() {
        val plan = planBatchNotification(summarizeBatch(listOf(transfer(TransferStatus.COMPLETED))))
        assertTrue(plan is TransferBatchNotificationPlan.Result)
    }

    @Test
    fun fullyCancelledBatchPlansNothing() {
        val plan = planBatchNotification(summarizeBatch(listOf(transfer(TransferStatus.CANCELLED))))
        assertEquals(TransferBatchNotificationPlan.Nothing, plan)
    }

    @Test
    fun finishedViewerOnlyDownloadBatchPlansNothing() {
        val plan = planBatchNotification(
            summarizeBatch(listOf(transfer(TransferStatus.COMPLETED, TransferDirection.DOWNLOAD, purpose = "OPEN_WITH"))),
        )
        assertEquals(TransferBatchNotificationPlan.Nothing, plan)
    }

    @Test
    fun singleSavedDownloadExposesTheOpenTarget() {
        val summary = summarizeBatch(
            listOf(
                transfer(
                    TransferStatus.COMPLETED,
                    TransferDirection.DOWNLOAD,
                    purpose = "SAVE_TO_DEVICE",
                    destinationUri = "content://saved",
                ),
            ),
        )!!
        assertNotNull(summary.singleSavedDownload)
        assertTrue(planBatchNotification(summary) is TransferBatchNotificationPlan.Result)
    }

    @Test
    fun multipleDownloadsDoNotExposeAnOpenTarget() {
        val saved = transfer(TransferStatus.COMPLETED, TransferDirection.DOWNLOAD, purpose = "SAVE_TO_DEVICE", destinationUri = "content://saved")
        val summary = summarizeBatch(listOf(saved, saved))!!
        assertNull(summary.singleSavedDownload)
    }

    @Test
    fun resultPartsOmitZeroCounts() {
        val summary = summarizeBatch(
            listOf(
                transfer(TransferStatus.COMPLETED),
                transfer(TransferStatus.COMPLETED),
                transfer(TransferStatus.FAILED),
            ),
        )!!
        assertEquals(
            listOf(BatchResultPart(BatchResultPartKind.TRANSFERRED, 2), BatchResultPart(BatchResultPartKind.FAILED, 1)),
            batchResultParts(summary),
        )
    }

    @Test
    fun resultPartsIncludeAlreadyExisted() {
        val summary = summarizeBatch(listOf(transfer(TransferStatus.ALREADY_EXISTED)))!!
        assertEquals(listOf(BatchResultPart(BatchResultPartKind.ALREADY_EXISTED, 1)), batchResultParts(summary))
    }

    @Test
    fun progressAndResultIdsDifferAndAreStablePerBatch() {
        val batchIds = listOf("a", "b", "11111111-2222-3333-4444-555555555555", "zzzzzzzz")
        batchIds.forEach { batchId ->
            assertNotEquals(TransferBatchNotificationIds.progress(batchId), TransferBatchNotificationIds.result(batchId))
            assertEquals(TransferBatchNotificationIds.progress(batchId), TransferBatchNotificationIds.progress(batchId))
            assertTrue(TransferBatchNotificationIds.progress(batchId) >= 0)
        }
        assertNotEquals(TransferBatchNotificationIds.progress("a"), TransferBatchNotificationIds.progress("b"))
    }
}
