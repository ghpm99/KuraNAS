package com.kuranas.android.feature.video.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.PlaybackStateDto
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
import org.junit.Assert.assertEquals
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
            PlaybackStateDto(playlistId = playlistId, videoId = 42, currentTime = currentTimeSeconds),
        )
    }

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
}
