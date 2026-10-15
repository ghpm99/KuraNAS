package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistItemDto
import com.kuranas.android.feature.video.data.VideoRepository
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
class VideoPlaylistViewModelTest {

    private val repository = mockk<VideoRepository>()
    private val playlistId = 7

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun itemOf(rawId: Int) = VideoPlaylistItemDto(video = VideoItemDto(rawId = rawId))

    private fun givenHeader(name: String = "Series") {
        coEvery { repository.getPlaylistHeader(playlistId) } returns AppResult.Success(VideoPlaylistDto(id = playlistId, name = name, itemCount = 3))
    }

    private fun givenPage(page: Int, hasNext: Boolean, vararg rawIds: Int) {
        coEvery { repository.getPlaylistItems(playlistId, page) } returns AppResult.Success(ListPage(rawIds.map(::itemOf), hasNext))
    }

    @Test
    fun `first page and header fill the state`() {
        givenHeader()
        givenPage(1, true, 1, 2)
        val viewModel = VideoPlaylistViewModel(repository)

        viewModel.load(playlistId)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertEquals("Series", state.playlist?.name)
        assertEquals(listOf("1", "2"), state.items.map { it.video.id })
        assertTrue(state.hasMore)
    }

    @Test
    fun `loadMore appends the next page without duplicates`() {
        givenHeader()
        givenPage(1, true, 1, 2)
        givenPage(2, false, 2, 3)
        val viewModel = VideoPlaylistViewModel(repository)
        viewModel.load(playlistId)

        viewModel.loadMore()

        val state = viewModel.state.value
        assertEquals(listOf("1", "2", "3"), state.items.map { it.video.id })
        assertFalse(state.hasMore)
    }

    @Test
    fun `loadMore stops requesting when there is no next page`() {
        givenHeader()
        givenPage(1, false, 1)
        val viewModel = VideoPlaylistViewModel(repository)
        viewModel.load(playlistId)

        viewModel.loadMore()

        coVerify(exactly = 0) { repository.getPlaylistItems(playlistId, 2) }
    }

    @Test
    fun `failed header keeps the server message`() {
        coEvery { repository.getPlaylistHeader(playlistId) } returns AppResult.Error("playlist not found")
        givenPage(1, false, 1)
        val viewModel = VideoPlaylistViewModel(repository)

        viewModel.load(playlistId)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertNull(state.playlist)
        assertEquals("playlist not found", state.error)
    }

    @Test
    fun `failed next page keeps the loaded items`() {
        givenHeader()
        givenPage(1, true, 1)
        coEvery { repository.getPlaylistItems(playlistId, 2) } returns AppResult.Error("boom")
        val viewModel = VideoPlaylistViewModel(repository)
        viewModel.load(playlistId)

        viewModel.loadMore()

        val state = viewModel.state.value
        assertEquals(listOf("1"), state.items.map { it.video.id })
        assertNull(state.error)
    }

    @Test
    fun `retry after a header error loads the playlist`() {
        coEvery { repository.getPlaylistHeader(playlistId) } returns AppResult.Error("boom")
        givenPage(1, false, 1)
        val viewModel = VideoPlaylistViewModel(repository)
        viewModel.load(playlistId)
        givenHeader()

        viewModel.retry()

        val state = viewModel.state.value
        assertNull(state.error)
        assertEquals("Series", state.playlist?.name)
    }
}
