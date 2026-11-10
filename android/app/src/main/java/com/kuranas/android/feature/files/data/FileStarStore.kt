package com.kuranas.android.feature.files.data

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class FileStarStore @Inject constructor() {

    private val _starredById = MutableStateFlow<Map<String, Boolean>>(emptyMap())
    val starredById: StateFlow<Map<String, Boolean>> = _starredById.asStateFlow()

    fun record(starredByFileId: Map<String, Boolean>) {
        _starredById.update { current -> current + starredByFileId }
    }

    fun set(fileId: String, isStarred: Boolean) {
        _starredById.update { current -> current + (fileId to isStarred) }
    }
}
