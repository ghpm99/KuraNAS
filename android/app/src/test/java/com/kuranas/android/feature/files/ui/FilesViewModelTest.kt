package com.kuranas.android.feature.files.ui

import androidx.lifecycle.SavedStateHandle
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredResponse
import io.mockk.coEvery
import io.mockk.coVerify
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
        coEvery { repository.getRootFiles() } returns
            AppResult.Success(listOf(FileItemDto(rawId = 7, name = "a.txt", starred = isStarred)))
        return FilesViewModel(repository, SavedStateHandle())
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
        coEvery { repository.getRootFiles() } returns AppResult.Success(emptyList())

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
}
