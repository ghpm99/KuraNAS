package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.NonCancellable
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import javax.inject.Inject

data class VideoPlayerStart(val streamUrl: String, val startPositionMs: Long)

data class PlayerProgress(val positionMs: Long, val durationMs: Long, val isPlaying: Boolean)

@HiltViewModel
class VideoPlayerViewModel @Inject constructor(private val repository: VideoRepository) : ViewModel() {
    private var currentVideoId: String? = null
    private var currentPlaylistId: Int? = null
    private var progressReportingJob: Job? = null

    fun getThumbnailUrl(videoId: String): String = runBlocking { repository.thumbnailUrl(videoId) }

    suspend fun openVideo(videoId: String): VideoPlayerStart {
        currentVideoId = videoId
        currentPlaylistId = null
        val streamUrl = repository.streamUrl(videoId)
        val playbackState = (repository.startPlayback(videoId) as? AppResult.Success)?.data
            ?: return VideoPlayerStart(streamUrl, startPositionMs = 0L)
        currentPlaylistId = playbackState.playlistId
        return VideoPlayerStart(streamUrl, secondsToMillis(playbackState.currentTime))
    }

    fun startProgressReporting(readProgress: () -> PlayerProgress) {
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
