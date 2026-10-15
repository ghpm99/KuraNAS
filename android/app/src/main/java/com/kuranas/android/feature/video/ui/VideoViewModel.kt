package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.video.data.VideoHomeCatalogDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistSection
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
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
    val playlistSections: List<VideoPlaylistSectionState> = emptyList(),
)

data class VideoPlaylistSectionState(
    val section: VideoPlaylistSection,
    val playlists: List<VideoPlaylistDto>,
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
            val (catalogResult, playlistSections) = coroutineScope {
                val catalog = async { repository.getHomeCatalog() }
                val sections = async { loadPlaylistSections() }
                catalog.await() to sections.await()
            }
            when (catalogResult) {
                is AppResult.Success -> _state.update {
                    it.copy(isLoading = false, isRefreshing = false, catalog = catalogResult.data, playlistSections = playlistSections)
                }
                is AppResult.Error -> _state.update {
                    it.copy(isLoading = false, isRefreshing = false, error = catalogResult.message, playlistSections = playlistSections)
                }
            }
        }
    }

    private suspend fun loadPlaylistSections(): List<VideoPlaylistSectionState> = coroutineScope {
        VideoPlaylistSection.entries
            .map { section -> async { section to repository.getPlaylistsBySection(section) } }
            .awaitAll()
            .mapNotNull { (section, result) ->
                val playlists = (result as? AppResult.Success)?.data.orEmpty()
                if (playlists.isEmpty()) null else VideoPlaylistSectionState(section, playlists)
            }
    }

    fun thumbnailUrl(videoId: String): String = runBlocking { repository.thumbnailUrl(videoId) }
}
