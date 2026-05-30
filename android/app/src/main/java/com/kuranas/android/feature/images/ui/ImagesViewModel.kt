package com.kuranas.android.feature.images.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.images.data.ImagesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ImagesUiState(
    val isLoading: Boolean = true,
    val images: List<FileItemDto> = emptyList(),
    val error: String? = null,
    val serverBaseUrl: String = "",
)

@HiltViewModel
class ImagesViewModel @Inject constructor(private val repository: ImagesRepository) : ViewModel() {

    private val _state = MutableStateFlow(ImagesUiState())
    val state: StateFlow<ImagesUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            when (val r = repository.getImages()) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, images = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun thumbnailUrl(id: String): String = runCatching {
        kotlinx.coroutines.runBlocking { repository.getThumbnailUrl(id) }
    }.getOrDefault("")
}
