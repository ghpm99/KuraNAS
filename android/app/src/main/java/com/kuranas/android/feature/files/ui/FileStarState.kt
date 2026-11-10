package com.kuranas.android.feature.files.ui

import com.kuranas.android.feature.files.data.FileStarToggler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class FileStarState(
    private val scope: CoroutineScope,
    private val toggler: FileStarToggler,
) {
    val starredById: StateFlow<Map<String, Boolean>> = toggler.starredById

    private val _actionErrorMessage = MutableStateFlow<String?>(null)
    val actionErrorMessage: StateFlow<String?> = _actionErrorMessage.asStateFlow()

    fun toggle(fileId: String, isStarredFallback: Boolean) {
        scope.launch {
            val errorMessage = toggler.toggle(fileId, isStarredFallback) ?: return@launch
            _actionErrorMessage.value = errorMessage
        }
    }

    fun dismissActionError() {
        _actionErrorMessage.value = null
    }
}
