package com.kuranas.android.feature.music.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.MusicPage
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.PlayerConnection
import io.mockk.coEvery
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class MusicAlbumViewModelTest {

    private val repository = mockk<MusicRepository>()
    private val player = mockk<PlayerConnection>(relaxed = true)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun tracksPage(hasNext: Boolean, vararg ids: Int) =
        AppResult.Success(MusicPage(ids.map { TrackDto(id = it) }, hasNext))

    @Test
    fun `load fetches the first page of the album tracks`() {
        coEvery { repository.getTracksByAlbum("album-key", 1) } returns tracksPage(true, 1, 2)
        val viewModel = MusicAlbumViewModel(repository, player)

        viewModel.load("album-key")

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertEquals(listOf(1, 2), state.items.map { it.id })
    }

    @Test
    fun `loadMore appends the next album page and play uses every loaded track`() {
        coEvery { repository.getTracksByAlbum("album-key", 1) } returns tracksPage(true, 1)
        coEvery { repository.getTracksByAlbum("album-key", 2) } returns tracksPage(false, 2)
        val viewModel = MusicAlbumViewModel(repository, player)
        viewModel.load("album-key")

        viewModel.loadMore()
        viewModel.play(viewModel.state.value.items.first())

        assertFalse(viewModel.state.value.hasMore)
        verify { player.play(TrackDto(id = 1), listOf(TrackDto(id = 1), TrackDto(id = 2))) }
    }

    @Test
    fun `failed first page ends loading with the error message`() {
        coEvery { repository.getTracksByAlbum("album-key", 1) } returns AppResult.Error("boom")
        val viewModel = MusicAlbumViewModel(repository, player)

        viewModel.load("album-key")

        assertFalse(viewModel.state.value.isLoading)
        assertEquals("boom", viewModel.state.value.error)
    }
}
