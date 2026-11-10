package com.kuranas.android.feature.trash.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.trash.data.TrashItemDto
import com.kuranas.android.feature.trash.data.TrashRepository
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
class TrashViewModelTest {

    private val repository = mockk<TrashRepository>()
    private val report = TrashItemDto(id = 1, name = "relatorio.pdf", originalPath = "/data/docs/relatorio.pdf")
    private val photos = TrashItemDto(id = 2, name = "fotos", originalPath = "/data/fotos", isDirectory = true)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun loadedViewModel(vararg trashedItems: TrashItemDto): TrashViewModel {
        coEvery { repository.getTrashItems(1) } returns AppResult.Success(ListPage(trashedItems.toList(), hasNext = false))
        return TrashViewModel(repository).also { it.loadTrash() }
    }

    private fun TrashViewModel.listedIds() = state.value.items.map { it.id }

    @Test
    fun `starts loading and lists the trash once asked`() {
        val viewModel = loadedViewModel(report, photos)

        assertFalse(viewModel.state.value.isLoading)
        assertEquals(listOf(1, 2), viewModel.listedIds())
    }

    @Test
    fun `a failed load shows the server message`() {
        coEvery { repository.getTrashItems(1) } returns AppResult.Error("lixeira indisponível")
        val viewModel = TrashViewModel(repository).also { it.loadTrash() }

        assertEquals("lixeira indisponível", viewModel.state.value.error)
        assertTrue(viewModel.state.value.items.isEmpty())
    }

    @Test
    fun `restore takes the item out of the list right away`() {
        val viewModel = loadedViewModel(report, photos)
        val pendingAnswer = CompletableDeferred<AppResult<Unit>>()
        coEvery { repository.restoreItem(1) } coAnswers { pendingAnswer.await() }

        viewModel.restore(report)

        assertEquals(listOf(2), viewModel.listedIds())
        pendingAnswer.complete(AppResult.Success(Unit))
        assertEquals(listOf(2), viewModel.listedIds())
    }

    @Test
    fun `a refused restore puts the item back in place and shows the server message`() {
        val viewModel = loadedViewModel(report, photos)
        coEvery { repository.restoreItem(1) } returns AppResult.Error("local ocupado")

        viewModel.restore(report)

        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertEquals("local ocupado", viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `delete forever waits for the confirmation`() {
        val viewModel = loadedViewModel(report)

        viewModel.requestDeleteForever(report)

        assertEquals(TrashConfirmation.DeleteForever(report), viewModel.state.value.pendingConfirmation)
        coVerify(exactly = 0) { repository.deleteItemForever(any()) }
    }

    @Test
    fun `cancelling the confirmation deletes nothing`() {
        val viewModel = loadedViewModel(report)

        viewModel.requestDeleteForever(report)
        viewModel.cancelConfirmation()
        viewModel.confirmPendingAction()

        assertNull(viewModel.state.value.pendingConfirmation)
        assertEquals(listOf(1), viewModel.listedIds())
        coVerify(exactly = 0) { repository.deleteItemForever(any()) }
    }

    @Test
    fun `confirmed delete forever removes the item`() {
        val viewModel = loadedViewModel(report, photos)
        coEvery { repository.deleteItemForever(2) } returns AppResult.Success(Unit)

        viewModel.requestDeleteForever(photos)
        viewModel.confirmPendingAction()

        coVerify(exactly = 1) { repository.deleteItemForever(2) }
        assertEquals(listOf(1), viewModel.listedIds())
        assertNull(viewModel.state.value.pendingConfirmation)
    }

    @Test
    fun `a refused delete forever puts the item back and shows the server message`() {
        val viewModel = loadedViewModel(report)
        coEvery { repository.deleteItemForever(1) } returns AppResult.Error("falha na lixeira")

        viewModel.requestDeleteForever(report)
        viewModel.confirmPendingAction()

        assertEquals(listOf(1), viewModel.listedIds())
        assertEquals("falha na lixeira", viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `confirmed empty trash calls the server and reloads the list`() {
        val viewModel = loadedViewModel(report, photos)
        coEvery { repository.emptyTrash() } coAnswers {
            coEvery { repository.getTrashItems(1) } returns AppResult.Success(ListPage(emptyList(), hasNext = false))
            AppResult.Success(Unit)
        }

        viewModel.requestEmptyTrash()
        assertEquals(TrashConfirmation.EmptyTrash, viewModel.state.value.pendingConfirmation)
        viewModel.confirmPendingAction()

        coVerify(exactly = 1) { repository.emptyTrash() }
        assertTrue(viewModel.state.value.items.isEmpty())
        assertFalse(viewModel.state.value.isEmptyingTrash)
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `a failed empty trash keeps what is left and shows the server message`() {
        val viewModel = loadedViewModel(report)
        coEvery { repository.emptyTrash() } returns AppResult.Error("falha ao esvaziar")

        viewModel.requestEmptyTrash()
        viewModel.confirmPendingAction()

        assertEquals(listOf(1), viewModel.listedIds())
        assertEquals("falha ao esvaziar", viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `dismissing clears the action error`() {
        val viewModel = loadedViewModel(report)
        coEvery { repository.restoreItem(1) } returns AppResult.Error("local ocupado")
        viewModel.restore(report)

        viewModel.dismissActionError()

        assertNull(viewModel.state.value.actionErrorMessage)
    }
}
