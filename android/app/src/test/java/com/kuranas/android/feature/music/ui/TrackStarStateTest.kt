package com.kuranas.android.feature.music.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileStarStore
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredResponse
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.data.TrackStarToggler
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class TrackStarStateTest {

    private val filesRepository = mockk<FilesRepository>()
    private val starStore = FileStarStore()
    private val toggler = TrackStarToggler(filesRepository, starStore)
    private val track = TrackDto(id = 7, starred = false)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun starState() = TrackStarState(TestScope(UnconfinedTestDispatcher()), toggler)

    @Test
    fun `toggle changes the star immediately and reconciles with the server`() {
        val pendingResponse = CompletableDeferred<AppResult<StarredResponse>>()
        coEvery { filesRepository.setFileStarred("7", true) } coAnswers { pendingResponse.await() }
        val stars = starState()

        stars.toggle(track)

        assertEquals(true, stars.starredById.value["7"])
        pendingResponse.complete(AppResult.Success(StarredResponse(7, true)))
        assertEquals(true, stars.starredById.value["7"])
        assertNull(stars.actionErrorMessage.value)
    }

    @Test
    fun `server response overrides the optimistic value`() {
        coEvery { filesRepository.setFileStarred("7", true) } returns AppResult.Success(StarredResponse(7, false))
        val stars = starState()

        stars.toggle(track)

        assertEquals(false, stars.starredById.value["7"])
    }

    @Test
    fun `error rolls back and exposes the server message verbatim`() {
        coEvery { filesRepository.setFileStarred("7", true) } returns AppResult.Error("falha do servidor")
        val stars = starState()

        stars.toggle(track)

        assertEquals(false, stars.starredById.value["7"])
        assertEquals("falha do servidor", stars.actionErrorMessage.value)
        stars.dismissActionError()
        assertNull(stars.actionErrorMessage.value)
    }

    @Test
    fun `list and player share the same star state through the store`() {
        coEvery { filesRepository.setFileStarred("7", true) } returns AppResult.Success(StarredResponse(7, true))
        val listStars = starState()
        val playerStars = starState()

        listStars.toggle(track)

        assertEquals(true, playerStars.starredById.value["7"])
    }

    @Test
    fun `toggle starts from the track flag when the store has no entry`() {
        coEvery { filesRepository.setFileStarred("7", false) } returns AppResult.Success(StarredResponse(7, false))
        val stars = starState()

        stars.toggle(TrackDto(id = 7, starred = true))

        assertEquals(false, stars.starredById.value["7"])
    }
}
