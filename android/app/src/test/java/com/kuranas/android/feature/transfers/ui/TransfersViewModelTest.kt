package com.kuranas.android.feature.transfers.ui

import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class TransfersViewModelTest {

    private val transfersStored = MutableStateFlow<List<TransferEntity>>(emptyList())
    private val transferDao = mockk<TransferDao> { every { observeAll() } returns transfersStored }
    private val transferRepository = mockk<TransferRepository>(relaxed = true)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun transferWith(
        id: Long,
        status: TransferStatus,
        direction: TransferDirection = TransferDirection.UPLOAD,
        totalBytes: Long = 100,
        transferredBytes: Long = 0,
        failureMessage: String? = null,
        relativePath: String? = null,
    ) = TransferEntity(
        id = id,
        batchId = "batch",
        direction = direction,
        displayName = "file-$id",
        targetFolderName = "Photos",
        relativePath = relativePath,
        totalBytes = totalBytes,
        transferredBytes = transferredBytes,
        status = status,
        failureMessage = failureMessage,
        createdAt = 0,
        updatedAt = 0,
    )

    @Test
    fun `groups transfers into in progress, completed and failed sections`() {
        val state = groupTransfers(
            listOf(
                transferWith(1, TransferStatus.QUEUED),
                transferWith(2, TransferStatus.RUNNING),
                transferWith(3, TransferStatus.WAITING_FOR_NETWORK),
                transferWith(4, TransferStatus.COMPLETED),
                transferWith(5, TransferStatus.ALREADY_EXISTED),
                transferWith(6, TransferStatus.CANCELLED),
                transferWith(7, TransferStatus.FAILED),
            ),
        )

        assertEquals(listOf(1L, 2L, 3L), state.inProgress.map { it.id })
        assertEquals(listOf(4L, 5L, 6L), state.completed.map { it.id })
        assertEquals(listOf(7L), state.failed.map { it.id })
    }

    @Test
    fun `progress fraction is known only when the total size is known`() {
        val half = transferWith(1, TransferStatus.RUNNING, totalBytes = 200, transferredBytes = 50).toTransferItem()
        val unknownSize = transferWith(2, TransferStatus.RUNNING, totalBytes = -1, transferredBytes = 50).toTransferItem()

        assertEquals(0.25f, half.progressFraction!!, 0.0001f)
        assertNull(unknownSize.progressFraction)
    }

    @Test
    fun `destination folder is shown for uploads only`() {
        val upload = transferWith(1, TransferStatus.QUEUED, TransferDirection.UPLOAD).toTransferItem()
        val download = transferWith(2, TransferStatus.QUEUED, TransferDirection.DOWNLOAD).toTransferItem()

        assertEquals("Photos", upload.destinationFolderName)
        assertNull(download.destinationFolderName)
    }

    @Test
    fun `upload destination appends the relative path when set`() {
        val upload = transferWith(1, TransferStatus.QUEUED, relativePath = "Celular/2026/11").toTransferItem()

        assertEquals("Photos/Celular/2026/11", upload.destinationFolderName)
    }

    @Test
    fun `cancel is offered to active items and retry to failed and cancelled ones`() {
        val running = transferWith(1, TransferStatus.RUNNING).toTransferItem()
        val failed = transferWith(2, TransferStatus.FAILED, failureMessage = "gone").toTransferItem()
        val cancelled = transferWith(3, TransferStatus.CANCELLED).toTransferItem()
        val completed = transferWith(4, TransferStatus.COMPLETED).toTransferItem()

        assertTrue(running.canCancel)
        assertFalse(running.canRetry)
        assertTrue(failed.canRetry)
        assertEquals("gone", failed.failureMessage)
        assertTrue(cancelled.canRetry)
        assertFalse(completed.canCancel || completed.canRetry)
    }

    @Test
    fun `clear finished is offered only when the completed section is not empty`() {
        assertFalse(groupTransfers(listOf(transferWith(1, TransferStatus.RUNNING))).hasFinishedItems)
        assertFalse(groupTransfers(listOf(transferWith(1, TransferStatus.FAILED))).hasFinishedItems)
        assertTrue(groupTransfers(listOf(transferWith(1, TransferStatus.COMPLETED))).hasFinishedItems)
        assertTrue(groupTransfers(emptyList()).isEmpty)
    }

    @Test
    fun `view model publishes the grouped stored transfers`() {
        val viewModel = TransfersViewModel(transferDao, transferRepository)
        val collection = CoroutineScope(Dispatchers.Unconfined).launch { viewModel.state.collect { } }
        transfersStored.value = listOf(transferWith(1, TransferStatus.RUNNING), transferWith(2, TransferStatus.FAILED))

        assertEquals(listOf(1L), viewModel.state.value.inProgress.map { it.id })
        assertEquals(listOf(2L), viewModel.state.value.failed.map { it.id })
        collection.cancel()
    }

    @Test
    fun `actions delegate to the repository`() {
        val viewModel = TransfersViewModel(transferDao, transferRepository)

        viewModel.cancel(1)
        viewModel.retry(2)
        viewModel.clearFinished()

        coVerify { transferRepository.cancel(1) }
        coVerify { transferRepository.retry(2) }
        coVerify { transferRepository.clearFinished() }
    }
}
