package com.kuranas.android.feature.music.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.AlbumDto
import com.kuranas.android.feature.music.data.ArtistDto
import com.kuranas.android.feature.music.data.FolderDto
import com.kuranas.android.feature.music.data.MusicPage
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.PlaylistDto
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
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class MusicViewModelTest {

    private val repository = mockk<MusicRepository>()
    private val player = mockk<PlayerConnection>(relaxed = true)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        coEvery { repository.getArtists(any()) } returns AppResult.Success(MusicPage(listOf(ArtistDto(key = "a")), false))
        coEvery { repository.getAlbums(any()) } returns AppResult.Success(MusicPage(listOf(AlbumDto(key = "b")), false))
        coEvery { repository.getPlaylists(any()) } returns AppResult.Success(MusicPage(listOf(PlaylistDto(id = 1)), false))
        coEvery { repository.getFolders(any()) } returns AppResult.Success(MusicPage(listOf(FolderDto(folder = "f")), false))
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun tracksPage(hasNext: Boolean, vararg ids: Int) =
        AppResult.Success(MusicPage(ids.map { TrackDto(id = it, name = "t$it") }, hasNext))

    @Test
    fun `loads the first page of every list on creation`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1, 2)

        val viewModel = MusicViewModel(repository, player)

        val tracks = viewModel.tracks.value
        assertFalse(tracks.isLoading)
        assertEquals(listOf(1, 2), tracks.items.map { it.id })
        assertTrue(tracks.hasMore)
        assertEquals(listOf("a"), viewModel.artists.value.items.map { it.key })
        assertEquals(listOf("f"), viewModel.folders.value.items.map { it.folder })
    }

    @Test
    fun `loadMoreTracks appends the next page`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1, 2)
        coEvery { repository.getAllTracks(2) } returns tracksPage(false, 3)
        val viewModel = MusicViewModel(repository, player)

        viewModel.loadMoreTracks()

        val tracks = viewModel.tracks.value
        assertEquals(listOf(1, 2, 3), tracks.items.map { it.id })
        assertFalse(tracks.hasMore)
        assertFalse(tracks.isLoadingMore)
    }

    @Test
    fun `loadMoreTracks stops requesting once the last page was reached`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1)
        coEvery { repository.getAllTracks(2) } returns tracksPage(false, 2)
        val viewModel = MusicViewModel(repository, player)

        viewModel.loadMoreTracks()
        viewModel.loadMoreTracks()
        viewModel.loadMoreTracks()

        coVerify(exactly = 1) { repository.getAllTracks(2) }
        coVerify(exactly = 0) { repository.getAllTracks(3) }
    }

    @Test
    fun `loadMoreTracks ignores duplicated tracks`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1, 2)
        coEvery { repository.getAllTracks(2) } returns tracksPage(false, 2, 3)
        val viewModel = MusicViewModel(repository, player)

        viewModel.loadMoreTracks()

        assertEquals(listOf(1, 2, 3), viewModel.tracks.value.items.map { it.id })
    }

    @Test
    fun `first page failure exposes the error`() {
        coEvery { repository.getAllTracks(1) } returns AppResult.Error("boom")

        val viewModel = MusicViewModel(repository, player)

        val tracks = viewModel.tracks.value
        assertFalse(tracks.isLoading)
        assertEquals("boom", tracks.error)
        assertTrue(tracks.items.isEmpty())
    }

    @Test
    fun `load more failure keeps loaded tracks and allows retry`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1)
        coEvery { repository.getAllTracks(2) } returns AppResult.Error("offline") andThen tracksPage(false, 2)
        val viewModel = MusicViewModel(repository, player)

        viewModel.loadMoreTracks()

        assertEquals(listOf(1), viewModel.tracks.value.items.map { it.id })
        assertNull(viewModel.tracks.value.error)
        assertFalse(viewModel.tracks.value.isLoadingMore)
        assertTrue(viewModel.tracks.value.hasMore)

        viewModel.loadMoreTracks()

        assertEquals(listOf(1, 2), viewModel.tracks.value.items.map { it.id })
    }

    @Test
    fun `refresh restarts from the first page`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(true, 1) andThen tracksPage(true, 9)
        coEvery { repository.getAllTracks(2) } returns tracksPage(false, 2)
        val viewModel = MusicViewModel(repository, player)
        viewModel.loadMoreTracks()

        viewModel.refresh()

        val tracks = viewModel.tracks.value
        assertEquals(listOf(9), tracks.items.map { it.id })
        assertFalse(tracks.isRefreshing)
        assertTrue(tracks.hasMore)
    }

    @Test
    fun `play queues the loaded tracks`() {
        coEvery { repository.getAllTracks(1) } returns tracksPage(false, 1, 2)
        val viewModel = MusicViewModel(repository, player)
        val loadedTracks = viewModel.tracks.value.items

        viewModel.play(loadedTracks[1], loadedTracks)

        verify { player.play(loadedTracks[1], loadedTracks) }
    }
}
