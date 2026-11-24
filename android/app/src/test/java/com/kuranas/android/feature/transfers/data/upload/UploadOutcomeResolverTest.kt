package com.kuranas.android.feature.transfers.data.upload

import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.TransferTransition
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class UploadOutcomeResolverTest {

    private val resolver = UploadOutcomeResolver(
        TransferStateMachine { 9_000L },
        UploadFailureMessages(sourceMissing = "source gone", rejectedWithoutDetail = "no detail"),
    )

    private val runningTransfer = TransferEntity(
        id = 1,
        batchId = "batch",
        direction = TransferDirection.UPLOAD,
        displayName = "photo.jpg",
        targetFolderName = "",
        totalBytes = 100,
        transferredBytes = 40,
        status = TransferStatus.RUNNING,
        createdAt = 1,
        updatedAt = 1,
    )

    private fun resolved(outcome: UploadOutcome): Pair<TransferEntity, Boolean> {
        val resolution = resolver.resolve(runningTransfer, outcome)
        return (resolution.transition as TransferTransition.Applied).transfer to resolution.shouldRetryLater
    }

    @Test
    fun completedMarksCompletedWithoutRetry() {
        val (transfer, shouldRetry) = resolved(UploadOutcome.Completed)
        assertEquals(TransferStatus.COMPLETED, transfer.status)
        assertFalse(shouldRetry)
    }

    @Test
    fun alreadyExistedMarksAlreadyExisted() {
        assertEquals(TransferStatus.ALREADY_EXISTED, resolved(UploadOutcome.AlreadyExisted).first.status)
    }

    @Test
    fun connectionLostWaitsForNetworkAndRetries() {
        val (transfer, shouldRetry) = resolved(UploadOutcome.ConnectionLost)
        assertEquals(TransferStatus.WAITING_FOR_NETWORK, transfer.status)
        assertTrue(shouldRetry)
    }

    @Test
    fun rejectedFailsWithServerMessageVerbatim() {
        val (transfer, shouldRetry) = resolved(UploadOutcome.Rejected("Sem espaço"))
        assertEquals(TransferStatus.FAILED, transfer.status)
        assertEquals("Sem espaço", transfer.failureMessage)
        assertFalse(shouldRetry)
    }

    @Test
    fun rejectedWithoutMessageUsesFallback() {
        assertEquals("no detail", resolved(UploadOutcome.Rejected(null)).first.failureMessage)
    }

    @Test
    fun sourceMissingFailsWithSourceMissingMessage() {
        val (transfer, _) = resolved(UploadOutcome.SourceMissing)
        assertEquals(TransferStatus.FAILED, transfer.status)
        assertEquals("source gone", transfer.failureMessage)
    }
}
