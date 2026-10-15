package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistItemDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class VideoPlaylistUiState(
    val isLoading: Boolean = true,
    val playlist: VideoPlaylistDto? = null,
    val items: List<VideoPlaylistItemDto> = emptyList(),
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val error: String? = null,
)

private data class PlaylistHeaderState(
    val isLoading: Boolean = true,
    val playlist: VideoPlaylistDto? = null,
    val error: String? = null,
)

@HiltViewModel
class VideoPlaylistViewModel @Inject constructor(val repository: VideoRepository) : ViewModel() {

    private val headerState = MutableStateFlow(PlaylistHeaderState())
    private val itemsLoader = PagedListLoader<VideoPlaylistItemDto>(viewModelScope) { it.video.id }

    val state: StateFlow<VideoPlaylistUiState> = combine(headerState, itemsLoader.state) { header, items ->
        VideoPlaylistUiState(
            isLoading = header.isLoading || items.isLoading,
            playlist = header.playlist,
            items = items.items,
            isLoadingMore = items.isLoadingMore,
            hasMore = items.hasMore,
            error = header.error ?: items.error,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, VideoPlaylistUiState())

    private var requestedPlaylistId: Int? = null

    fun load(id: Int) {
        requestedPlaylistId = id
        headerState.value = PlaylistHeaderState()
        itemsLoader.load { page -> repository.getPlaylistItems(id, page) }
        viewModelScope.launch {
            when (val result = repository.getPlaylistHeader(id)) {
                is AppResult.Success -> headerState.update { PlaylistHeaderState(isLoading = false, playlist = result.data) }
                is AppResult.Error -> headerState.update { PlaylistHeaderState(isLoading = false, error = result.message) }
            }
        }
    }

    fun loadMore() = itemsLoader.loadMore()

    fun retry() {
        val playlistId = requestedPlaylistId ?: return
        load(playlistId)
    }
}
