package com.kuranas.android.feature.images.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.images.data.ImageLibraryItemDto
import com.kuranas.android.feature.images.data.ImageLibraryPageDto
import com.kuranas.android.feature.images.data.ImagesRepository
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
class ImagesViewModelTest {

    private val repository = mockk<ImagesRepository>()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun images(vararg fileIds: Int) = fileIds.map { ImageLibraryItemDto(fileId = it, name = "photo$it.jpg") }

    private fun page(nextCursor: String, vararg fileIds: Int) = ImageLibraryPageDto(
        items = images(*fileIds),
        nextCursor = nextCursor,
        hasNext = nextCursor.isNotEmpty(),
    )

    private fun givenFirstPage(page: ImageLibraryPageDto) {
        coEvery { repository.getImagesPage(null) } returns AppResult.Success(page)
    }

    @Test
    fun `loads the first page on creation`() {
        givenFirstPage(page("cursor-1", 3, 2, 1))

        val viewModel = ImagesViewModel(repository)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertEquals(listOf("3", "2", "1"), state.images.map { it.id })
        assertTrue(state.hasMore)
        assertEquals("cursor-1", state.nextCursor)
    }

    @Test
    fun `loadMore appends the next page using the cursor`() {
        givenFirstPage(page("cursor-1", 3, 2))
        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Success(page("", 1))
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()

        val state = viewModel.state.value
        assertEquals(listOf("3", "2", "1"), state.images.map { it.id })
        assertFalse(state.hasMore)
        assertFalse(state.isLoadingMore)
    }

    @Test
    fun `loadMore walks every page until the last one`() {
        givenFirstPage(page("cursor-1", 6, 5))
        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Success(page("cursor-2", 4, 3))
        coEvery { repository.getImagesPage("cursor-2") } returns AppResult.Success(page("", 2, 1))
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()
        viewModel.loadMore()
        viewModel.loadMore()

        assertEquals(listOf("6", "5", "4", "3", "2", "1"), viewModel.state.value.images.map { it.id })
        coVerify(exactly = 1) { repository.getImagesPage("cursor-1") }
        coVerify(exactly = 1) { repository.getImagesPage("cursor-2") }
    }

    @Test
    fun `loadMore does nothing when there is no next page`() {
        givenFirstPage(page("", 1))
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()

        coVerify(exactly = 1) { repository.getImagesPage(any()) }
    }

    @Test
    fun `loadMore ignores calls while a page is already loading`() {
        givenFirstPage(page("cursor-1", 2, 1))
        val pendingPage = CompletableDeferred<AppResult<ImageLibraryPageDto>>()
        coEvery { repository.getImagesPage("cursor-1") } coAnswers { pendingPage.await() }
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()
        viewModel.loadMore()

        assertTrue(viewModel.state.value.isLoadingMore)
        pendingPage.complete(AppResult.Success(page("", 0)))
        coVerify(exactly = 1) { repository.getImagesPage("cursor-1") }
    }

    @Test
    fun `loadMore drops duplicated images across pages`() {
        givenFirstPage(page("cursor-1", 3, 2))
        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Success(page("", 2, 1))
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()

        assertEquals(listOf("3", "2", "1"), viewModel.state.value.images.map { it.id })
    }

    @Test
    fun `loadMore failure keeps loaded images and allows retrying`() {
        givenFirstPage(page("cursor-1", 2, 1))
        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Error("offline")
        val viewModel = ImagesViewModel(repository)

        viewModel.loadMore()

        val failedState = viewModel.state.value
        assertEquals(2, failedState.images.size)
        assertNull(failedState.error)
        assertFalse(failedState.isLoadingMore)
        assertTrue(failedState.hasMore)

        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Success(page("", 0))
        viewModel.loadMore()
        assertEquals(3, viewModel.state.value.images.size)
    }

    @Test
    fun `first page failure exposes the error`() {
        coEvery { repository.getImagesPage(null) } returns AppResult.Error("offline")

        val viewModel = ImagesViewModel(repository)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertEquals("offline", state.error)
        assertTrue(state.images.isEmpty())
    }

    @Test
    fun `refresh restarts from the first page and replaces loaded images`() {
        givenFirstPage(page("cursor-1", 3, 2))
        coEvery { repository.getImagesPage("cursor-1") } returns AppResult.Success(page("", 1))
        val viewModel = ImagesViewModel(repository)
        viewModel.loadMore()

        givenFirstPage(page("cursor-9", 9, 8))
        viewModel.refresh()

        val state = viewModel.state.value
        assertFalse(state.isRefreshing)
        assertEquals(listOf("9", "8"), state.images.map { it.id })
        assertEquals("cursor-9", state.nextCursor)
        assertTrue(state.hasMore)
    }

    @Test
    fun `refresh discards a page that was still loading`() {
        givenFirstPage(page("cursor-1", 2, 1))
        val stalePage = CompletableDeferred<AppResult<ImageLibraryPageDto>>()
        coEvery { repository.getImagesPage("cursor-1") } coAnswers { stalePage.await() }
        val viewModel = ImagesViewModel(repository)
        viewModel.loadMore()

        givenFirstPage(page("cursor-9", 9))
        viewModel.refresh()
        stalePage.complete(AppResult.Success(page("", 7)))

        assertEquals(listOf("9"), viewModel.state.value.images.map { it.id })
    }
}
