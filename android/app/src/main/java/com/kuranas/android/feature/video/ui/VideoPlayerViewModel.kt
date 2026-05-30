package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.runBlocking
import javax.inject.Inject

@HiltViewModel
class VideoPlayerViewModel @Inject constructor(private val repository: VideoRepository) : ViewModel() {
    fun getStreamUrl(videoId: String): String = runBlocking { repository.streamUrl(videoId) }
    fun getThumbnailUrl(videoId: String): String = runBlocking { repository.thumbnailUrl(videoId) }
}
