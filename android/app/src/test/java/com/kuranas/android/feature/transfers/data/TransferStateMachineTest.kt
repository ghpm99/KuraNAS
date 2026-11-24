package com.kuranas.android.feature.transfers.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TransferStateMachineTest {

    private val clockMillis = 5_000L
    private val machine = TransferStateMachine { clockMillis }

    private fun transferIn(
        status: TransferStatus,
        totalBytes: Long = 100,
        transferredBytes: Long = 10,
    ) = TransferEntity(
        id = 1,
        batchId = "batch",
        direction = TransferDirection.UPLOAD,
        displayName = "photo.jpg",
        targetFolderName = "Root",
        totalBytes = totalBytes,
        transferredBytes = transferredBytes,
        uploadId = "upload-1",
        status = status,
        failureMessage = "boom",
        createdAt = 1,
        updatedAt = 1,
    )

    private fun applied(transition: TransferTransition): TransferEntity {
        assertTrue("expected Applied but was $transition", transition is TransferTransition.Applied)
        return (transition as TransferTransition.Applied).transfer
    }

    private fun assertRejected(transition: TransferTransition, original: TransferEntity) {
        assertEquals(TransferTransition.Rejected(original, original.status), transition)
    }

    private fun allStatusesExcept(vararg allowed: TransferStatus) =
        TransferStatus.entries.filter { it !in allowed }

    @Test
    fun startMovesQueuedAndWaitingToRunning() {
        listOf(TransferStatus.QUEUED, TransferStatus.WAITING_FOR_NETWORK).forEach { status ->
            val next = applied(machine.start(transferIn(status)))
            assertEquals(TransferStatus.RUNNING, next.status)
            assertEquals(clockMillis, next.updatedAt)
        }
    }

    @Test
    fun startIsRejectedFromOtherStatuses() {
        allStatusesExcept(TransferStatus.QUEUED, TransferStatus.WAITING_FOR_NETWORK).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.start(transfer), transfer)
        }
    }

    @Test
    fun progressUpdatesBytesOnlyWhileRunning() {
        val next = applied(machine.progress(transferIn(TransferStatus.RUNNING), 55))
        assertEquals(55, next.transferredBytes)
        assertEquals(TransferStatus.RUNNING, next.status)
        assertEquals(clockMillis, next.updatedAt)
        allStatusesExcept(TransferStatus.RUNNING).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.progress(transfer, 55), transfer)
        }
    }

    @Test
    fun waitForNetworkMovesRunningToWaiting() {
        val next = applied(machine.waitForNetwork(transferIn(TransferStatus.RUNNING)))
        assertEquals(TransferStatus.WAITING_FOR_NETWORK, next.status)
        allStatusesExcept(TransferStatus.RUNNING).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.waitForNetwork(transfer), transfer)
        }
    }

    @Test
    fun completeSetsTransferredBytesToTotal() {
        val next = applied(machine.complete(transferIn(TransferStatus.RUNNING, totalBytes = 100, transferredBytes = 40)))
        assertEquals(TransferStatus.COMPLETED, next.status)
        assertEquals(100, next.transferredBytes)
    }

    @Test
    fun completeKeepsTransferredBytesWhenTotalUnknown() {
        val next = applied(machine.complete(transferIn(TransferStatus.RUNNING, totalBytes = -1, transferredBytes = 40)))
        assertEquals(40, next.transferredBytes)
    }

    @Test
    fun completeIsRejectedFromNonRunning() {
        allStatusesExcept(TransferStatus.RUNNING).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.complete(transfer), transfer)
        }
    }

    @Test
    fun alreadyExistedMovesRunningOnly() {
        val next = applied(machine.alreadyExisted(transferIn(TransferStatus.RUNNING)))
        assertEquals(TransferStatus.ALREADY_EXISTED, next.status)
        allStatusesExcept(TransferStatus.RUNNING).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.alreadyExisted(transfer), transfer)
        }
    }

    @Test
    fun failStoresMessageFromActiveStatuses() {
        listOf(TransferStatus.QUEUED, TransferStatus.RUNNING, TransferStatus.WAITING_FOR_NETWORK).forEach { status ->
            val next = applied(machine.fail(transferIn(status), "network down"))
            assertEquals(TransferStatus.FAILED, next.status)
            assertEquals("network down", next.failureMessage)
        }
    }

    @Test
    fun failIsRejectedFromTerminalStatuses() {
        listOf(
            TransferStatus.COMPLETED,
            TransferStatus.ALREADY_EXISTED,
            TransferStatus.FAILED,
            TransferStatus.CANCELLED,
        ).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.fail(transfer, "x"), transfer)
        }
    }

    @Test
    fun cancelMovesAnyActiveStatusToCancelled() {
        listOf(TransferStatus.QUEUED, TransferStatus.RUNNING, TransferStatus.WAITING_FOR_NETWORK).forEach { status ->
            assertEquals(TransferStatus.CANCELLED, applied(machine.cancel(transferIn(status))).status)
        }
    }

    @Test
    fun cancelIsRejectedFromTerminalStatuses() {
        listOf(
            TransferStatus.COMPLETED,
            TransferStatus.ALREADY_EXISTED,
            TransferStatus.FAILED,
            TransferStatus.CANCELLED,
        ).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.cancel(transfer), transfer)
        }
    }

    @Test
    fun retryRequeuesFailedAndCancelledKeepingProgressAndUploadId() {
        listOf(TransferStatus.FAILED, TransferStatus.CANCELLED).forEach { status ->
            val next = applied(machine.retry(transferIn(status, transferredBytes = 30)))
            assertEquals(TransferStatus.QUEUED, next.status)
            assertNull(next.failureMessage)
            assertEquals("upload-1", next.uploadId)
            assertEquals(30, next.transferredBytes)
            assertEquals(clockMillis, next.updatedAt)
        }
    }

    @Test
    fun retryIsRejectedFromOtherStatuses() {
        allStatusesExcept(TransferStatus.FAILED, TransferStatus.CANCELLED).forEach { status ->
            val transfer = transferIn(status)
            assertRejected(machine.retry(transfer), transfer)
        }
    }
}
