package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoHomeCatalogDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import javax.inject.Inject

data class VideoUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val catalog: VideoHomeCatalogDto? = null,
    val error: String? = null,
)

@HiltViewModel
class VideoViewModel @Inject constructor(private val repository: VideoRepository) : ViewModel() {

    private val _state = MutableStateFlow(VideoUiState())
    val state: StateFlow<VideoUiState> = _state.asStateFlow()

    init { load() }

    /** Recarrega o catálogo exibindo o indicador de pull-to-refresh (sem spinner de tela cheia). */
    fun refresh() = load(refreshing = true)

    fun load(refreshing: Boolean = false) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = !refreshing, isRefreshing = refreshing, error = null) }
            when (val r = repository.getHomeCatalog()) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, isRefreshing = false, catalog = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, isRefreshing = false, error = r.message) }
            }
        }
    }

    fun thumbnailUrl(videoId: String): String = runBlocking { repository.thumbnailUrl(videoId) }
}
