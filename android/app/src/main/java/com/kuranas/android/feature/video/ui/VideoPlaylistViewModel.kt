package com.kuranas.android.feature.video.ui

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class VideoPlaylistViewModel @Inject constructor(val repository: VideoRepository) : ViewModel() {
    var playlist by mutableStateOf<VideoPlaylistDto?>(null)
        private set

    fun load(id: Int) {
        viewModelScope.launch {
            when (val r = repository.getPlaylistById(id)) {
                is AppResult.Success -> playlist = r.data
                is AppResult.Error -> playlist = VideoPlaylistDto()
            }
        }
    }
}
