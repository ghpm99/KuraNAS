package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoHomeCatalogDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistSection
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
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class VideoViewModelTest {

    private val repository = mockk<VideoRepository>()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        coEvery { repository.getHomeCatalog() } returns AppResult.Success(VideoHomeCatalogDto())
        VideoPlaylistSection.entries.forEach { section ->
            coEvery { repository.getPlaylistsBySection(section, any(), any()) } returns AppResult.Success(emptyList())
        }
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun givenSection(section: VideoPlaylistSection, result: AppResult<List<VideoPlaylistDto>>) {
        coEvery { repository.getPlaylistsBySection(section, any(), any()) } returns result
    }

    @Test
    fun `loaded sections appear in the state in frontend order`() {
        givenSection(VideoPlaylistSection.Folders, AppResult.Success(listOf(VideoPlaylistDto(id = 3))))
        givenSection(VideoPlaylistSection.Series, AppResult.Success(listOf(VideoPlaylistDto(id = 1), VideoPlaylistDto(id = 2))))

        val viewModel = VideoViewModel(repository)

        val sections = viewModel.state.value.playlistSections
        assertEquals(listOf(VideoPlaylistSection.Series, VideoPlaylistSection.Folders), sections.map { it.section })
        assertEquals(listOf(1, 2), sections[0].playlists.map { it.id })
    }

    @Test
    fun `empty section does not appear`() {
        givenSection(VideoPlaylistSection.Movies, AppResult.Success(emptyList()))

        val viewModel = VideoViewModel(repository)

        assertEquals(emptyList<VideoPlaylistSectionState>(), viewModel.state.value.playlistSections)
    }

    @Test
    fun `failing section does not break the catalog or the other sections`() {
        givenSection(VideoPlaylistSection.Series, AppResult.Error("boom"))
        givenSection(VideoPlaylistSection.Clips, AppResult.Success(listOf(VideoPlaylistDto(id = 5))))

        val viewModel = VideoViewModel(repository)

        val state = viewModel.state.value
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertNotNull(state.catalog)
        assertEquals(listOf(VideoPlaylistSection.Clips), state.playlistSections.map { it.section })
    }
}
