package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoRepository
import io.mockk.coEvery
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

    @Test
    fun `successful load fills the playlist`() {
        coEvery { repository.getPlaylistById(playlistId) } returns AppResult.Success(VideoPlaylistDto(id = playlistId, name = "Series"))
        val viewModel = VideoPlaylistViewModel(repository)

        viewModel.load(playlistId)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertEquals("Series", state.playlist?.name)
    }

    @Test
    fun `failed load keeps the server message and no playlist`() {
        coEvery { repository.getPlaylistById(playlistId) } returns AppResult.Error("playlist not found")
        val viewModel = VideoPlaylistViewModel(repository)

        viewModel.load(playlistId)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertNull(state.playlist)
        assertEquals("playlist not found", state.error)
    }

    @Test
    fun `retry after an error loads the playlist`() {
        coEvery { repository.getPlaylistById(playlistId) } returns AppResult.Error("boom")
        val viewModel = VideoPlaylistViewModel(repository)
        viewModel.load(playlistId)
        coEvery { repository.getPlaylistById(playlistId) } returns AppResult.Success(VideoPlaylistDto(id = playlistId, name = "Series"))

        viewModel.retry()

        val state = viewModel.state.value
        assertNull(state.error)
        assertEquals("Series", state.playlist?.name)
    }
}
