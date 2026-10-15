package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.PlaybackSessionDto
import com.kuranas.android.feature.video.data.PlaybackStateDto
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistItemDto
import com.kuranas.android.feature.video.data.VideoRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import kotlinx.coroutines.flow.first
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class VideoPlayerViewModelTest {

    private val repository = mockk<VideoRepository>()
    private val videoId = "42"
    private val playlistId = 9
    private val intervalMs = VideoPlayerViewModel.PROGRESS_REPORT_INTERVAL_MS

    @Before
    fun setUp() {
        Dispatchers.setMain(StandardTestDispatcher())
        coEvery { repository.streamUrl(videoId) } returns "http://nas/stream/42"
        coEvery { repository.updatePlaybackState(any(), any(), any(), any(), any()) } returns AppResult.Success(Unit)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun givenStartSucceeds(currentTimeSeconds: Double) {
        coEvery { repository.startPlayback(videoId) } returns AppResult.Success(
            PlaybackSessionDto(
                playbackState = PlaybackStateDto(playlistId = playlistId, videoId = 42, currentTime = currentTimeSeconds),
            ),
        )
    }

    private fun sessionAt(currentVideoId: Int, currentTimeSeconds: Double = 0.0, playlistVideoIds: List<Int> = listOf(41, 42, 43)) =
        PlaybackSessionDto(
            playlist = VideoPlaylistDto(
                id = playlistId,
                items = playlistVideoIds.map { VideoPlaylistItemDto(video = VideoItemDto(rawId = it)) },
            ),
            playbackState = PlaybackStateDto(
                playlistId = playlistId,
                videoId = currentVideoId,
                currentTime = currentTimeSeconds,
            ),
        )

    private fun givenStartReturns(session: PlaybackSessionDto) {
        coEvery { repository.startPlayback(videoId) } returns AppResult.Success(session)
    }

    private val playing = PlayerProgress(positionMs = 20_000, durationMs = 100_000, isPlaying = true)

    private fun givenStartFails() {
        coEvery { repository.startPlayback(videoId) } returns AppResult.Error("falhou")
    }

    private suspend fun openedViewModel(): VideoPlayerViewModel {
        val viewModel = VideoPlayerViewModel(repository)
        viewModel.openVideo(videoId)
        return viewModel
    }

    private fun TestScope.settle() = runCurrent()

    @Test
    fun `successful start exposes the saved position in milliseconds`() = runTest {
        givenStartSucceeds(currentTimeSeconds = 125.5)

        val start = VideoPlayerViewModel(repository).openVideo(videoId)

        assertEquals(125_500L, start.startPositionMs)
        assertEquals("http://nas/stream/42", start.streamUrl)
    }

    @Test
    fun `failed start begins at zero and still returns the stream url`() = runTest {
        givenStartFails()

        val start = VideoPlayerViewModel(repository).openVideo(videoId)

        assertEquals(0L, start.startPositionMs)
        assertEquals("http://nas/stream/42", start.streamUrl)
    }

    @Test
    fun `periodic report sends ids position duration and playing state after the interval`() = runTest {
        givenStartSucceeds(0.0)
        val viewModel = openedViewModel()

        viewModel.startProgressReporting { PlayerProgress(positionMs = 30_000, durationMs = 600_000, isPlaying = true) }
        advanceTimeBy(intervalMs - 1)
        settle()
        coVerify(exactly = 0) { repository.updatePlaybackState(any(), any(), any(), any(), any()) }

        advanceTimeBy(1)
        settle()
        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, playlistId, 30.0, 600.0, false) }
        viewModel.stopProgressReporting()
    }

    @Test
    fun `periodic report is skipped while the player is paused`() = runTest {
        givenStartSucceeds(0.0)
        val viewModel = openedViewModel()

        viewModel.startProgressReporting { PlayerProgress(positionMs = 1_000, durationMs = 10_000, isPlaying = false) }
        advanceTimeBy(intervalMs * 2)
        settle()

        coVerify(exactly = 0) { repository.updatePlaybackState(any(), any(), any(), any(), any()) }
        viewModel.stopProgressReporting()
    }

    @Test
    fun `pause or exit reports immediately as paused`() = runTest {
        givenStartSucceeds(0.0)
        val viewModel = openedViewModel()

        viewModel.reportProgressNow(PlayerProgress(positionMs = 12_000, durationMs = 60_000, isPlaying = false))
        settle()

        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, playlistId, 12.0, 60.0, true) }
    }

    @Test
    fun `unknown duration is not sent`() = runTest {
        givenStartSucceeds(0.0)
        val viewModel = openedViewModel()

        viewModel.reportProgressNow(PlayerProgress(positionMs = 5_000, durationMs = Long.MIN_VALUE + 1, isPlaying = false))
        settle()

        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, playlistId, 5.0, null, true) }
    }

    @Test
    fun `report after failed start sends only the video id`() = runTest {
        givenStartFails()
        val viewModel = openedViewModel()

        viewModel.reportProgressNow(PlayerProgress(positionMs = 5_000, durationMs = 10_000, isPlaying = false))
        settle()

        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, null, 5.0, 10.0, true) }
    }

    @Test
    fun `middle of a three item playlist allows previous and next`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42))

        val viewModel = openedViewModel()

        assertTrue(viewModel.navigationAvailability.value.canGoPrevious)
        assertTrue(viewModel.navigationAvailability.value.canGoNext)
    }

    @Test
    fun `first item allows only next`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42, playlistVideoIds = listOf(42, 43, 44)))

        val viewModel = openedViewModel()

        assertFalse(viewModel.navigationAvailability.value.canGoPrevious)
        assertTrue(viewModel.navigationAvailability.value.canGoNext)
    }

    @Test
    fun `last item allows only previous`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42, playlistVideoIds = listOf(40, 41, 42)))

        val viewModel = openedViewModel()

        assertTrue(viewModel.navigationAvailability.value.canGoPrevious)
        assertFalse(viewModel.navigationAvailability.value.canGoNext)
    }

    @Test
    fun `go to next reports progress then exposes destination media and reports with the new video id`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42))
        coEvery { repository.nextVideo() } returns AppResult.Success(sessionAt(currentVideoId = 43, currentTimeSeconds = 7.5))
        coEvery { repository.streamUrl("43") } returns "http://nas/stream/43"
        val viewModel = openedViewModel()
        viewModel.startProgressReporting { playing }

        viewModel.goToNext()
        settle()
        val change = viewModel.mediaChanges.first()

        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, playlistId, 20.0, 100.0, true) }
        coVerify(exactly = 1) { repository.nextVideo() }
        assertEquals("http://nas/stream/43", change.streamUrl)
        assertEquals(7_500L, change.startPositionMs)

        viewModel.reportProgressNow(PlayerProgress(positionMs = 8_000, durationMs = 50_000, isPlaying = false))
        settle()
        coVerify(exactly = 1) { repository.updatePlaybackState("43", playlistId, 8.0, 50.0, true) }
        viewModel.stopProgressReporting()
    }

    @Test
    fun `go to previous requests the previous video`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42))
        coEvery { repository.previousVideo() } returns AppResult.Success(sessionAt(currentVideoId = 41))
        coEvery { repository.streamUrl("41") } returns "http://nas/stream/41"
        val viewModel = openedViewModel()

        viewModel.goToPrevious()
        settle()

        assertEquals("http://nas/stream/41", viewModel.mediaChanges.first().streamUrl)
        assertFalse(viewModel.navigationAvailability.value.canGoPrevious)
    }

    @Test
    fun `backend returning the same video does not request a media change`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42, playlistVideoIds = listOf(41, 42)))
        coEvery { repository.nextVideo() } returns AppResult.Success(sessionAt(currentVideoId = 42, playlistVideoIds = listOf(41, 42)))
        val viewModel = openedViewModel()

        viewModel.goToNext()
        settle()

        coVerify(exactly = 1) { repository.streamUrl(videoId) }
    }

    @Test
    fun `next failure exposes the server message and keeps the current video`() = runTest {
        givenStartReturns(sessionAt(currentVideoId = 42))
        coEvery { repository.nextVideo() } returns AppResult.Error("sem proximo")
        val viewModel = openedViewModel()

        viewModel.goToNext()
        settle()
        viewModel.reportProgressNow(PlayerProgress(positionMs = 5_000, durationMs = 10_000, isPlaying = false))
        settle()

        assertEquals("sem proximo", viewModel.navigationError.value)
        assertTrue(viewModel.navigationAvailability.value.canGoNext)
        coVerify(exactly = 1) { repository.updatePlaybackState(videoId, playlistId, 5.0, 10.0, true) }
    }

    @Test
    fun `opening with a playlist starts playback inside that playlist and uses its session for navigation`() = runTest {
        coEvery { repository.startPlayback(videoId, playlistId) } returns AppResult.Success(sessionAt(currentVideoId = 42))
        val viewModel = VideoPlayerViewModel(repository)

        viewModel.openVideo(videoId, playlistId)

        coVerify(exactly = 1) { repository.startPlayback(videoId, playlistId) }
        coVerify(exactly = 0) { repository.startPlayback(videoId, null) }
        assertEquals(VideoNavigationAvailability(canGoPrevious = true, canGoNext = true), viewModel.navigationAvailability.value)
    }

    @Test
    fun `opening without a playlist starts playback without playlist id`() = runTest {
        givenStartSucceeds(currentTimeSeconds = 0.0)

        VideoPlayerViewModel(repository).openVideo(videoId)

        coVerify(exactly = 1) { repository.startPlayback(videoId, null) }
    }

    @Test
    fun `failed start with playlist falls back to start without playlist`() = runTest {
        coEvery { repository.startPlayback(videoId, playlistId) } returns AppResult.Error("fora da playlist")
        givenStartSucceeds(currentTimeSeconds = 30.0)

        val start = VideoPlayerViewModel(repository).openVideo(videoId, playlistId)

        coVerify(exactly = 1) { repository.startPlayback(videoId, playlistId) }
        coVerify(exactly = 1) { repository.startPlayback(videoId, null) }
        assertEquals(30_000L, start.startPositionMs)
    }
}
