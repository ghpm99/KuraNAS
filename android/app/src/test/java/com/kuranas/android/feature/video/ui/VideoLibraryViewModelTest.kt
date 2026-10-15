package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class VideoLibraryViewModelTest {

    private val repository = mockk<VideoRepository>()
    private val dispatcher = UnconfinedTestDispatcher()

    @Before
    fun setUp() {
        Dispatchers.setMain(dispatcher)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun givenPage(searchQuery: String, page: Int, hasNext: Boolean, vararg rawIds: Int) {
        coEvery { repository.getLibraryVideos(searchQuery, page) } returns
            AppResult.Success(ListPage(rawIds.map { VideoItemDto(rawId = it) }, hasNext))
    }

    @Test
    fun `initial load fills the first page`() {
        givenPage("", 1, true, 1, 2)

        val state = VideoLibraryViewModel(repository).state.value

        assertFalse(state.isLoading)
        assertNull(state.error)
        assertEquals(listOf("1", "2"), state.videos.map { it.id })
        assertTrue(state.hasMore)
    }

    @Test
    fun `loadMore appends the next page without duplicates`() {
        givenPage("", 1, true, 1, 2)
        givenPage("", 2, false, 2, 3)
        val viewModel = VideoLibraryViewModel(repository)

        viewModel.loadMore()

        val state = viewModel.state.value
        assertEquals(listOf("1", "2", "3"), state.videos.map { it.id })
        assertFalse(state.hasMore)
    }

    @Test
    fun `fast typing triggers a single search after the debounce and restarts at page one`() {
        givenPage("", 1, false, 1)
        givenPage("matrix", 1, false, 9)
        val viewModel = VideoLibraryViewModel(repository)

        viewModel.onSearchQueryChange("m")
        dispatcher.scheduler.advanceTimeBy(200)
        viewModel.onSearchQueryChange("ma")
        dispatcher.scheduler.advanceTimeBy(200)
        viewModel.onSearchQueryChange("matrix")
        dispatcher.scheduler.advanceTimeBy(399)
        dispatcher.scheduler.runCurrent()
        coVerify(exactly = 0) { repository.getLibraryVideos("matrix", 1) }

        dispatcher.scheduler.advanceTimeBy(1)
        dispatcher.scheduler.runCurrent()

        coVerify(exactly = 1) { repository.getLibraryVideos("matrix", 1) }
        coVerify(exactly = 0) { repository.getLibraryVideos("m", any()) }
        coVerify(exactly = 0) { repository.getLibraryVideos("ma", any()) }
        assertEquals(listOf("9"), viewModel.state.value.videos.map { it.id })
        assertEquals("matrix", viewModel.state.value.searchQuery)
    }

    @Test
    fun `server error is shown as received`() {
        coEvery { repository.getLibraryVideos("", 1) } returns AppResult.Error("library unavailable")

        val state = VideoLibraryViewModel(repository).state.value

        assertFalse(state.isLoading)
        assertEquals("library unavailable", state.error)
    }
}
