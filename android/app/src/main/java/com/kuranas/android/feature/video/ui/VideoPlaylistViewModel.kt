package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class VideoPlaylistUiState(
    val isLoading: Boolean = true,
    val playlist: VideoPlaylistDto? = null,
    val error: String? = null,
)

@HiltViewModel
class VideoPlaylistViewModel @Inject constructor(val repository: VideoRepository) : ViewModel() {

    private val _state = MutableStateFlow(VideoPlaylistUiState())
    val state: StateFlow<VideoPlaylistUiState> = _state.asStateFlow()

    private var requestedPlaylistId: Int? = null

    fun load(id: Int) {
        requestedPlaylistId = id
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            when (val result = repository.getPlaylistById(id)) {
                is AppResult.Success -> _state.update { VideoPlaylistUiState(isLoading = false, playlist = result.data) }
                is AppResult.Error -> _state.update { VideoPlaylistUiState(isLoading = false, error = result.message) }
            }
        }
    }

    fun retry() {
        val playlistId = requestedPlaylistId ?: return
        load(playlistId)
    }
}
