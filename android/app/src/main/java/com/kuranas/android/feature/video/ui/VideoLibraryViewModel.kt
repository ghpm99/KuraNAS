package com.kuranas.android.feature.video.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

private const val SEARCH_DEBOUNCE_MILLIS = 400L

data class VideoLibraryUiState(
    val searchQuery: String = "",
    val isLoading: Boolean = true,
    val videos: List<VideoItemDto> = emptyList(),
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class VideoLibraryViewModel @Inject constructor(private val repository: VideoRepository) : ViewModel() {

    private val searchQuery = MutableStateFlow("")
    private val videosLoader = PagedListLoader<VideoItemDto>(viewModelScope) { it.id }
    private var searchJob: Job? = null

    val state: StateFlow<VideoLibraryUiState> = combine(searchQuery, videosLoader.state) { query, videos ->
        VideoLibraryUiState(
            searchQuery = query,
            isLoading = videos.isLoading,
            videos = videos.items,
            isLoadingMore = videos.isLoadingMore,
            hasMore = videos.hasMore,
            error = videos.error,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, VideoLibraryUiState())

    init {
        loadFirstPage()
    }

    fun onSearchQueryChange(query: String) {
        searchQuery.value = query
        searchJob?.cancel()
        searchJob = viewModelScope.launch {
            delay(SEARCH_DEBOUNCE_MILLIS)
            loadFirstPage()
        }
    }

    fun loadMore() = videosLoader.loadMore()

    fun retry() = loadFirstPage()

    private fun loadFirstPage() {
        val query = searchQuery.value
        videosLoader.load { page -> repository.getLibraryVideos(query, page) }
    }
}
