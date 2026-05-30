package com.kuranas.android.feature.music.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.AlbumDto
import com.kuranas.android.feature.music.data.ArtistDto
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.PlaylistDto
import com.kuranas.android.feature.music.data.TrackDto
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class MusicTab { TRACKS, ARTISTS, ALBUMS, PLAYLISTS }

data class MusicUiState(
    val tab: MusicTab = MusicTab.ARTISTS,
    val isLoading: Boolean = true,
    val tracks: List<TrackDto> = emptyList(),
    val artists: List<ArtistDto> = emptyList(),
    val albums: List<AlbumDto> = emptyList(),
    val playlists: List<PlaylistDto> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class MusicViewModel @Inject constructor(private val repository: MusicRepository) : ViewModel() {

    private val _state = MutableStateFlow(MusicUiState())
    val state: StateFlow<MusicUiState> = _state.asStateFlow()

    init { loadAll() }

    fun selectTab(tab: MusicTab) {
        _state.update { it.copy(tab = tab) }
    }

    private fun loadAll() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true) }
            val artists = repository.getArtists()
            val albums = repository.getAlbums()
            val playlists = repository.getPlaylists()
            _state.update {
                it.copy(
                    isLoading = false,
                    artists = (artists as? AppResult.Success)?.data ?: emptyList(),
                    albums = (albums as? AppResult.Success)?.data ?: emptyList(),
                    playlists = (playlists as? AppResult.Success)?.data ?: emptyList(),
                    error = (artists as? AppResult.Error)?.message,
                )
            }
        }
    }

    fun loadTracks() {
        viewModelScope.launch {
            when (val r = repository.getAllTracks()) {
                is AppResult.Success -> _state.update { it.copy(tracks = r.data) }
                is AppResult.Error -> _state.update { it.copy(error = r.message) }
            }
        }
    }
}
