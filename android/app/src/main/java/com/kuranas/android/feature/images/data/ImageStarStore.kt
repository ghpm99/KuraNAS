package com.kuranas.android.feature.images.data

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ImageStarStore @Inject constructor() {

    private val _starredById = MutableStateFlow<Map<String, Boolean>>(emptyMap())
    val starredById: StateFlow<Map<String, Boolean>> = _starredById.asStateFlow()

    fun record(images: List<ImageLibraryItemDto>) {
        _starredById.update { current -> current + images.associate { it.id to it.starred } }
    }

    fun set(imageId: String, isStarred: Boolean) {
        _starredById.update { current -> current + (imageId to isStarred) }
    }
}
