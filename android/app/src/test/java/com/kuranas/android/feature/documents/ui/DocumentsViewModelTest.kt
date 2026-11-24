package com.kuranas.android.feature.documents.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.documents.data.DocumentDto
import com.kuranas.android.feature.documents.data.DocumentsRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
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
class DocumentsViewModelTest {

    private val repository = mockk<DocumentsRepository>()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Success(3)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun document(id: Int) = DocumentDto(id = id, name = "doc$id.pdf", format = ".pdf")

    private fun DocumentsViewModel.listedIds() = state.value.documents.map { it.id }

    @Test
    fun `lists the first page and the total from the count`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1), document(2)), hasNext = true))

        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }

        assertFalse(viewModel.state.value.isLoading)
        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertTrue(viewModel.state.value.hasMore)
        assertEquals(3, viewModel.state.value.totalDocuments)
    }

    @Test
    fun `startIfNeeded does not reload after a rotation`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        val viewModel = DocumentsViewModel(repository)

        viewModel.startIfNeeded()
        viewModel.startIfNeeded()

        coVerify(exactly = 1) { repository.getDocuments(1, any(), any()) }
        assertEquals(listOf(1), viewModel.listedIds())
    }

    @Test
    fun `loadMore appends the next page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = true))
        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Success(ListPage(listOf(document(2)), hasNext = false))
        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }

        viewModel.loadMore()

        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertFalse(viewModel.state.value.hasMore)
    }

    @Test
    fun `a next page error keeps the items and retry loads that page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = true))
        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Error("falha na página")
        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }

        viewModel.loadMore()

        assertEquals(listOf(1), viewModel.listedIds())
        assertEquals("falha na página", viewModel.state.value.loadMoreError)
        assertNull(viewModel.state.value.error)

        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Success(ListPage(listOf(document(2)), hasNext = false))
        viewModel.retryLoadMore()

        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertNull(viewModel.state.value.loadMoreError)
    }

    @Test
    fun `refresh restarts from the first page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(5), document(6)), hasNext = false))
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Success(2)

        viewModel.refresh()

        assertEquals(listOf(5, 6), viewModel.listedIds())
        assertEquals(2, viewModel.state.value.totalDocuments)
        assertFalse(viewModel.state.value.isRefreshing)
    }

    @Test
    fun `a refused first load shows the server message verbatim`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Error("acesso negado a este IP", 403)

        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }

        assertEquals("acesso negado a este IP", viewModel.state.value.error)
        assertTrue(viewModel.state.value.documents.isEmpty())
    }

    @Test
    fun `a failed count hides the total without breaking the list`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Error("sem total")

        val viewModel = DocumentsViewModel(repository).also { it.startIfNeeded() }

        assertNull(viewModel.state.value.totalDocuments)
        assertEquals(listOf(1), viewModel.listedIds())
    }
}
