package com.kuranas.android.feature.music.ui

import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.PlayerConnection
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MusicFolderViewModel @Inject constructor(
    private val repository: MusicRepository,
    private val player: PlayerConnection,
) : ViewModel() {
    private val _tracks = mutableStateOf<List<TrackDto>?>(null)
    val tracks: State<List<TrackDto>?> = _tracks

    fun load(key: String) {
        viewModelScope.launch {
            when (val r = repository.getTracksByFolder(key)) {
                is AppResult.Success -> _tracks.value = r.data
                is AppResult.Error -> _tracks.value = emptyList()
            }
        }
    }

    fun play(track: TrackDto) = player.play(track, _tracks.value ?: emptyList())
}
