package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.PlaybackSessionDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.NonCancellable
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import javax.inject.Inject

data class VideoPlayerStart(val streamUrl: String, val startPositionMs: Long)

data class PlayerProgress(val positionMs: Long, val durationMs: Long, val isPlaying: Boolean)

data class VideoNavigationAvailability(val canGoPrevious: Boolean = false, val canGoNext: Boolean = false)

@HiltViewModel
class VideoPlayerViewModel @Inject constructor(private val repository: VideoRepository) : ViewModel() {
    private var currentVideoId: String? = null
    private var currentPlaylistId: Int? = null
    private var progressReportingJob: Job? = null
    private var progressReader: (() -> PlayerProgress)? = null
    private var playlistVideoIds: List<String> = emptyList()

    private val navigationAvailabilityState = MutableStateFlow(VideoNavigationAvailability())
    val navigationAvailability: StateFlow<VideoNavigationAvailability> = navigationAvailabilityState.asStateFlow()

    private val navigationErrorState = MutableStateFlow<String?>(null)
    val navigationError: StateFlow<String?> = navigationErrorState.asStateFlow()

    private val mediaChangeChannel = Channel<VideoPlayerStart>(Channel.BUFFERED)
    val mediaChanges = mediaChangeChannel.receiveAsFlow()

    fun getThumbnailUrl(videoId: String): String = runBlocking { repository.thumbnailUrl(videoId) }

    suspend fun openVideo(videoId: String): VideoPlayerStart {
        currentVideoId = videoId
        currentPlaylistId = null
        playlistVideoIds = emptyList()
        navigationErrorState.value = null
        val streamUrl = repository.streamUrl(videoId)
        val session = (repository.startPlayback(videoId) as? AppResult.Success)?.data
        if (session == null) {
            refreshNavigationAvailability()
            return VideoPlayerStart(streamUrl, startPositionMs = 0L)
        }
        applySession(session)
        return VideoPlayerStart(streamUrl, secondsToMillis(session.playbackState.currentTime))
    }

    fun goToNext() = navigate { repository.nextVideo() }

    fun goToPrevious() = navigate { repository.previousVideo() }

    private fun navigate(requestShift: suspend () -> AppResult<PlaybackSessionDto>) {
        viewModelScope.launch {
            navigationErrorState.value = null
            progressReader?.let { readProgress -> sendProgress(readProgress().copy(isPlaying = false)) }
            val previousVideoId = currentVideoId
            when (val shift = requestShift()) {
                is AppResult.Error -> navigationErrorState.value = shift.message
                is AppResult.Success -> {
                    applySession(shift.data)
                    if (currentVideoId == previousVideoId) return@launch
                    val destinationId = currentVideoId ?: return@launch
                    mediaChangeChannel.send(
                        VideoPlayerStart(
                            streamUrl = repository.streamUrl(destinationId),
                            startPositionMs = secondsToMillis(shift.data.playbackState.currentTime),
                        ),
                    )
                }
            }
        }
    }

    private fun applySession(session: PlaybackSessionDto) {
        currentPlaylistId = session.playbackState.playlistId
        session.playbackState.videoId?.let { currentVideoId = it.toString() }
        playlistVideoIds = session.playlist.items.map { it.video.id }
        refreshNavigationAvailability()
    }

    private fun refreshNavigationAvailability() {
        val currentIndex = playlistVideoIds.indexOf(currentVideoId)
        navigationAvailabilityState.value = VideoNavigationAvailability(
            canGoPrevious = currentIndex > 0,
            canGoNext = currentIndex >= 0 && currentIndex < playlistVideoIds.lastIndex,
        )
    }

    fun startProgressReporting(readProgress: () -> PlayerProgress) {
        progressReader = readProgress
        progressReportingJob?.cancel()
        progressReportingJob = viewModelScope.launch {
            while (true) {
                delay(PROGRESS_REPORT_INTERVAL_MS)
                val progress = readProgress()
                if (progress.isPlaying) sendProgress(progress)
            }
        }
    }

    fun stopProgressReporting() {
        progressReportingJob?.cancel()
        progressReportingJob = null
        progressReader = null
    }

    fun reportProgressNow(progress: PlayerProgress) {
        viewModelScope.launch { sendProgress(progress) }
    }

    private suspend fun sendProgress(progress: PlayerProgress) {
        val videoId = currentVideoId ?: return
        withContext(NonCancellable) {
            repository.updatePlaybackState(
                videoId = videoId,
                playlistId = currentPlaylistId,
                currentTimeSeconds = millisToSeconds(progress.positionMs),
                durationSeconds = progress.durationMs.takeIf { it > 0 }?.let(::millisToSeconds),
                isPaused = !progress.isPlaying,
            )
        }
    }

    private fun secondsToMillis(seconds: Double): Long = (seconds * MILLIS_PER_SECOND).toLong().coerceAtLeast(0L)

    private fun millisToSeconds(millis: Long): Double = millis / MILLIS_PER_SECOND

    companion object {
        const val PROGRESS_REPORT_INTERVAL_MS = 5_000L
        private const val MILLIS_PER_SECOND = 1000.0
    }
}
