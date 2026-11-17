package com.kuranas.android.feature.music.ui

import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileStarStore
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.files.data.FileStarToggler
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.PlayerConnection
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class MusicArtistViewModelTest {

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
        AppResult.Success(ListPage(ids.map { TrackDto(id = it) }, hasNext))

    @Test
    fun `load more failure keeps loaded tracks and retry requests the same page`() {
        coEvery { repository.getTracksByArtist("artist-key", 1) } returns tracksPage(true, 1)
        coEvery { repository.getTracksByArtist("artist-key", 2) } returns AppResult.Error("offline") andThen tracksPage(false, 2)
        val viewModel = MusicArtistViewModel(repository, player, FileStarToggler(mockk(), FileStarStore()))
        viewModel.load("artist-key")

        viewModel.loadMore()

        assertEquals(listOf(1), viewModel.state.value.items.map { it.id })
        assertEquals("offline", viewModel.state.value.loadMoreError)

        viewModel.retryLoadMore()

        assertEquals(listOf(1, 2), viewModel.state.value.items.map { it.id })
        assertNull(viewModel.state.value.loadMoreError)
        coVerify(exactly = 2) { repository.getTracksByArtist("artist-key", 2) }
    }
}
