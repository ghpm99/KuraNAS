package com.kuranas.android.feature.images.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.images.data.ImageLibraryItemDto
import com.kuranas.android.feature.images.data.ImageLibraryPageDto
import com.kuranas.android.feature.images.data.ImagesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ImagesUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val isLoadingMore: Boolean = false,
    val images: List<ImageLibraryItemDto> = emptyList(),
    val nextCursor: String = "",
    val hasMore: Boolean = false,
    val error: String? = null,
    val serverBaseUrl: String = "",
)

@HiltViewModel
class ImagesViewModel @Inject constructor(private val repository: ImagesRepository) : ViewModel() {

    private val _state = MutableStateFlow(ImagesUiState())
    val state: StateFlow<ImagesUiState> = _state.asStateFlow()

    private var pageLoadJob: Job? = null

    init { load() }

    fun load() {
        restartFromFirstPage(
            onStart = { it.copy(isLoading = true, isRefreshing = false, isLoadingMore = false, error = null) },
            onError = { current, message -> current.copy(isLoading = false, error = message) },
        )
    }

    fun refresh() {
        restartFromFirstPage(
            onStart = { it.copy(isRefreshing = true, isLoadingMore = false, error = null) },
            onError = { current, message -> current.copy(isRefreshing = false, error = message) },
        )
    }

    fun loadMore() {
        val current = _state.value
        val isBusy = current.isLoading || current.isRefreshing || current.isLoadingMore
        if (!current.hasMore || isBusy) return

        _state.update { it.copy(isLoadingMore = true) }
        pageLoadJob = viewModelScope.launch {
            when (val result = repository.getImagesPage(current.nextCursor)) {
                is AppResult.Success -> _state.update { it.withAppendedPage(result.data) }
                is AppResult.Error -> _state.update { it.copy(isLoadingMore = false) }
            }
        }
    }

    fun thumbnailUrl(id: String): String = runCatching {
        kotlinx.coroutines.runBlocking { repository.getThumbnailUrl(id) }
    }.getOrDefault("")

    private fun restartFromFirstPage(
        onStart: (ImagesUiState) -> ImagesUiState,
        onError: (ImagesUiState, String) -> ImagesUiState,
    ) {
        pageLoadJob?.cancel()
        _state.update(onStart)
        pageLoadJob = viewModelScope.launch {
            when (val result = repository.getImagesPage(cursor = null)) {
                is AppResult.Success -> _state.update { it.withFirstPage(result.data) }
                is AppResult.Error -> _state.update { onError(it, result.message) }
            }
        }
    }

    private fun ImagesUiState.withFirstPage(page: ImageLibraryPageDto) = copy(
        isLoading = false,
        isRefreshing = false,
        images = page.items,
        nextCursor = page.nextCursor,
        hasMore = page.hasNext && page.nextCursor.isNotEmpty(),
    )

    private fun ImagesUiState.withAppendedPage(page: ImageLibraryPageDto) = copy(
        isLoadingMore = false,
        images = (images + page.items).distinctBy { it.fileId },
        nextCursor = page.nextCursor,
        hasMore = page.hasNext && page.nextCursor.isNotEmpty(),
    )
}
