package com.kuranas.android.feature.images.ui

import androidx.lifecycle.ViewModel
import com.kuranas.android.feature.images.data.ImagesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.runBlocking
import javax.inject.Inject

@HiltViewModel
class ImageViewerViewModel @Inject constructor(val repository: ImagesRepository) : ViewModel() {

    fun getBlobUrl(fileId: String): String = runBlocking { repository.getBlobUrl(fileId) }
    fun getThumbnailUrl(fileId: String): String = runBlocking { repository.getThumbnailUrl(fileId) }
}
