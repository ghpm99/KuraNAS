package com.kuranas.android.feature.images.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.images.data.ImageStarStore
import com.kuranas.android.feature.images.data.ImagesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import javax.inject.Inject

@HiltViewModel
class ImageViewerViewModel @Inject constructor(
    val repository: ImagesRepository,
    private val filesRepository: FilesRepository,
    private val starStore: ImageStarStore,
) : ViewModel() {

    val starredById: StateFlow<Map<String, Boolean>> = starStore.starredById

    private val _actionErrorMessage = MutableStateFlow<String?>(null)
    val actionErrorMessage: StateFlow<String?> = _actionErrorMessage.asStateFlow()

    fun getBlobUrl(fileId: String): String = runBlocking { repository.getBlobUrl(fileId) }
    fun getThumbnailUrl(fileId: String): String = runBlocking { repository.getThumbnailUrl(fileId) }

    fun toggleStarred(fileId: String) {
        val previousIsStarred = starredById.value[fileId] ?: false
        starStore.set(fileId, !previousIsStarred)
        viewModelScope.launch {
            when (val result = filesRepository.setFileStarred(fileId, !previousIsStarred)) {
                is AppResult.Success -> starStore.set(fileId, result.data.starred)
                is AppResult.Error -> {
                    starStore.set(fileId, previousIsStarred)
                    _actionErrorMessage.update { result.message }
                }
            }
        }
    }

    fun dismissActionError() {
        _actionErrorMessage.value = null
    }
}
