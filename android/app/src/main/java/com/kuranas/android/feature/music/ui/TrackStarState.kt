package com.kuranas.android.feature.music.ui

import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.data.TrackStarToggler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class TrackStarState(
    private val scope: CoroutineScope,
    private val toggler: TrackStarToggler,
) {
    val starredById: StateFlow<Map<String, Boolean>> = toggler.starredById

    private val _actionErrorMessage = MutableStateFlow<String?>(null)
    val actionErrorMessage: StateFlow<String?> = _actionErrorMessage.asStateFlow()

    fun toggle(track: TrackDto) {
        scope.launch {
            val errorMessage = toggler.toggle(track) ?: return@launch
            _actionErrorMessage.value = errorMessage
        }
    }

    fun dismissActionError() {
        _actionErrorMessage.value = null
    }
}
