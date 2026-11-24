package com.kuranas.android.feature.files.ui

import androidx.lifecycle.SavedStateHandle
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredResponse
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.verify
import io.mockk.mockk
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
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
class FilesViewModelTest {

    private val repository = mockk<FilesRepository>()
    private val dispatcher = UnconfinedTestDispatcher()

    @Before
    fun setUp() {
        Dispatchers.setMain(dispatcher)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun viewModelWithFile(isStarred: Boolean): FilesViewModel {
        coEvery { repository.getRootFiles(any()) } returns
            AppResult.Success(ListPage(listOf(FileItemDto(rawId = 7, name = "a.txt", starred = isStarred)), false))
        return FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())
    }

    private fun FilesViewModel.isFileStarred() = state.value.files.single().starred

    @Test
    fun `tap flips the star immediately while the request is pending`() {
        val viewModel = viewModelWithFile(isStarred = false)
        val pendingResponse = CompletableDeferred<AppResult<StarredResponse>>()
        coEvery { repository.setFileStarred("7", true) } coAnswers { pendingResponse.await() }

        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        assertTrue(viewModel.isFileStarred())
        pendingResponse.complete(AppResult.Success(StarredResponse(7, true)))
    }

    @Test
    fun `success keeps the star returned by the server`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.setFileStarred("7", true) } returns AppResult.Success(StarredResponse(7, true))

        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        assertTrue(viewModel.isFileStarred())
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `success reconciles with the starred value returned by the server`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.setFileStarred("7", true) } returns AppResult.Success(StarredResponse(7, false))

        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        assertFalse(viewModel.isFileStarred())
    }

    @Test
    fun `unstarring a starred file requests starred false`() {
        val viewModel = viewModelWithFile(isStarred = true)
        coEvery { repository.setFileStarred("7", false) } returns AppResult.Success(StarredResponse(7, false))

        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        coVerify { repository.setFileStarred("7", false) }
        assertFalse(viewModel.isFileStarred())
    }

    @Test
    fun `error rolls back the star and exposes the server message`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.setFileStarred("7", true) } returns AppResult.Error("Arquivo não encontrado", 404)

        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        assertFalse(viewModel.isFileStarred())
        assertEquals("Arquivo não encontrado", viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `dismissActionError clears the message`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.setFileStarred("7", true) } returns AppResult.Error("falhou")
        viewModel.toggleFileStarred(viewModel.state.value.files.single())

        viewModel.dismissActionError()

        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `delete waits for confirmation before calling the server`() {
        val viewModel = viewModelWithFile(isStarred = false)
        val file = viewModel.state.value.files.single()

        viewModel.requestDelete(file)

        assertEquals(file, viewModel.state.value.fileAwaitingDeleteConfirmation)
        coVerify(exactly = 0) { repository.deleteFile(any()) }
    }

    @Test
    fun `cancelling the confirmation keeps the file and calls nothing`() {
        val viewModel = viewModelWithFile(isStarred = false)

        viewModel.requestDelete(viewModel.state.value.files.single())
        viewModel.cancelDelete()

        assertNull(viewModel.state.value.fileAwaitingDeleteConfirmation)
        assertEquals(1, viewModel.state.value.files.size)
        coVerify(exactly = 0) { repository.deleteFile(any()) }
    }

    @Test
    fun `confirmed delete removes the file from the list`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.deleteFile("7") } returns AppResult.Success(Unit)
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(emptyList(), false))

        viewModel.requestDelete(viewModel.state.value.files.single())
        viewModel.confirmDelete()

        coVerify(exactly = 1) { repository.deleteFile("7") }
        assertTrue(viewModel.state.value.files.isEmpty())
        assertNull(viewModel.state.value.fileAwaitingDeleteConfirmation)
    }

    @Test
    fun `failed delete keeps the file and shows the server message`() {
        val viewModel = viewModelWithFile(isStarred = false)
        coEvery { repository.deleteFile("7") } returns AppResult.Error("mensagem do servidor")

        viewModel.requestDelete(viewModel.state.value.files.single())
        viewModel.confirmDelete()

        assertEquals(1, viewModel.state.value.files.size)
        assertEquals("mensagem do servidor", viewModel.state.value.actionErrorMessage)
    }

    private fun fileNamed(rawId: Int) = FileItemDto(rawId = rawId, name = "file-$rawId.txt")

    @Test
    fun `second page appends its items to the folder list`() {
        coEvery { repository.getRootFiles(1) } returns AppResult.Success(ListPage(listOf(fileNamed(1)), true))
        coEvery { repository.getRootFiles(2) } returns AppResult.Success(ListPage(listOf(fileNamed(2)), false))
        val viewModel = FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())

        viewModel.loadMore()

        assertEquals(listOf("1", "2"), viewModel.state.value.files.map { it.id })
        assertFalse(viewModel.state.value.hasMore)
    }

    @Test
    fun `end of list does not request another page`() {
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(listOf(fileNamed(1)), false))
        val viewModel = FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())

        viewModel.loadMore()

        coVerify(exactly = 1) { repository.getRootFiles(any()) }
    }

    @Test
    fun `next page failure keeps the items and retry asks for the same page`() {
        coEvery { repository.getRootFiles(1) } returns AppResult.Success(ListPage(listOf(fileNamed(1)), true))
        coEvery { repository.getRootFiles(2) } returns AppResult.Error("falhou")
        val viewModel = FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())

        viewModel.loadMore()

        assertEquals(1, viewModel.state.value.files.size)
        assertEquals("falhou", viewModel.state.value.loadMoreError)

        coEvery { repository.getRootFiles(2) } returns AppResult.Success(ListPage(listOf(fileNamed(2)), false))
        viewModel.retryLoadMore()

        coVerify(exactly = 2) { repository.getRootFiles(2) }
        assertEquals(2, viewModel.state.value.files.size)
        assertNull(viewModel.state.value.loadMoreError)
    }

    @Test
    fun `opening a folder restarts from the first page`() {
        coEvery { repository.getRootFiles(1) } returns AppResult.Success(ListPage(listOf(fileNamed(1)), true))
        coEvery { repository.getRootFiles(2) } returns AppResult.Success(ListPage(listOf(fileNamed(2)), true))
        coEvery { repository.getChildrenById("9", 1) } returns AppResult.Success(ListPage(listOf(fileNamed(3)), false))
        val viewModel = FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())
        viewModel.loadMore()

        viewModel.openFolder(FileItemDto(rawId = 9, name = "docs", type = FileItemDto.TYPE_DIRECTORY))

        assertEquals(listOf("3"), viewModel.state.value.files.map { it.id })
        assertEquals(listOf("docs" to "9"), viewModel.state.value.breadcrumb)
        coVerify(exactly = 1) { repository.getChildrenById("9", 1) }
    }

    @Test
    fun `creating a folder reloads the current folder`() {
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(listOf(fileNamed(1)), false))
        coEvery { repository.createFolder("novo", null) } returns AppResult.Success(Unit)
        val viewModel = FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), FakeUploadCompletionSource(), SavedStateHandle())
        coEvery { repository.getRootFiles(any()) } returns
            AppResult.Success(ListPage(listOf(fileNamed(1), fileNamed(2)), false))

        viewModel.createFolder("novo")

        assertEquals(2, viewModel.state.value.files.size)
    }

    @Test
    fun `selected files are queued for download in list order and the selection is cleared`() {
        val queue = mockk<DownloadQueue>(relaxed = true)
        val folder = FileItemDto(rawId = 1, name = "pasta", type = FileItemDto.TYPE_DIRECTORY)
        val first = FileItemDto(rawId = 2, name = "primeiro arquivo.pdf", format = ".pdf", size = 10)
        val second = FileItemDto(rawId = 3, name = "foto.png", format = ".png", size = 20)
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(listOf(folder, first, second), false))
        val viewModel = FilesViewModel(repository, queue, FakeUploadCompletionSource(), SavedStateHandle())

        viewModel.toggleSelection(second)
        viewModel.toggleSelection(folder)
        viewModel.toggleSelection(first)
        viewModel.downloadSelected()

        coVerify {
            queue.enqueue(
                listOf(
                    DownloadRequest("2", "primeiro arquivo.pdf", 10),
                    DownloadRequest("3", "foto.png", 20),
                ),
            )
        }
        assertTrue(viewModel.state.value.selectedFileIds.isEmpty())
        assertEquals(2, viewModel.state.value.queuedDownloadsCount)
    }

    @Test
    fun `toggling a selected file again unselects it and an empty selection queues nothing`() {
        val queue = mockk<DownloadQueue>(relaxed = true)
        val file = FileItemDto(rawId = 2, name = "a.txt")
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(listOf(file), false))
        val viewModel = FilesViewModel(repository, queue, FakeUploadCompletionSource(), SavedStateHandle())

        viewModel.toggleSelection(file)
        assertTrue(viewModel.state.value.isSelecting)
        viewModel.toggleSelection(file)
        viewModel.downloadSelected()

        assertFalse(viewModel.state.value.isSelecting)
        coVerify(exactly = 0) { queue.enqueue(any()) }
    }

    private fun uploadTransfer(id: Long, status: TransferStatus, folderId: Long?) = TransferEntity(
        id = id,
        batchId = "batch",
        direction = TransferDirection.UPLOAD,
        displayName = "f$id",
        targetFolderId = folderId,
        targetFolderName = "folder",
        status = status,
        createdAt = 0,
        updatedAt = 0,
    )

    @Test
    fun `upload completing in the open root folder refreshes the listing once for a burst`() {
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(emptyList(), false))
        val uploads = FakeUploadCompletionSource()
        FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), uploads, SavedStateHandle())

        uploads.transfers.value = listOf(uploadTransfer(1, TransferStatus.RUNNING, null), uploadTransfer(2, TransferStatus.RUNNING, null))
        uploads.transfers.value = listOf(uploadTransfer(1, TransferStatus.COMPLETED, null), uploadTransfer(2, TransferStatus.RUNNING, null))
        uploads.transfers.value = listOf(uploadTransfer(1, TransferStatus.COMPLETED, null), uploadTransfer(2, TransferStatus.COMPLETED, null))
        dispatcher.scheduler.advanceTimeBy(UPLOAD_REFRESH_DEBOUNCE_MILLIS + 1)

        coVerify(exactly = 2) { repository.getRootFiles(any()) }
    }

    @Test
    fun `upload completing in another folder does not refresh the listing`() {
        coEvery { repository.getRootFiles(any()) } returns AppResult.Success(ListPage(emptyList(), false))
        val uploads = FakeUploadCompletionSource()
        FilesViewModel(repository, mockk<DownloadQueue>(relaxed = true), uploads, SavedStateHandle())

        uploads.transfers.value = listOf(uploadTransfer(1, TransferStatus.RUNNING, 9))
        uploads.transfers.value = listOf(uploadTransfer(1, TransferStatus.COMPLETED, 9))
        dispatcher.scheduler.advanceTimeBy(UPLOAD_REFRESH_DEBOUNCE_MILLIS + 1)

        coVerify(exactly = 1) { repository.getRootFiles(any()) }
    }
}
