package com.kuranas.android.feature.music.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.AlbumDto
import com.kuranas.android.feature.music.data.ArtistDto
import com.kuranas.android.feature.music.data.FolderDto
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.PlaylistDto
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.PlayerConnection
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class MusicTab { TRACKS, ARTISTS, ALBUMS, PLAYLISTS, FOLDERS }

data class MusicUiState(
    val tab: MusicTab = MusicTab.TRACKS,
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val tracks: List<TrackDto> = emptyList(),
    val artists: List<ArtistDto> = emptyList(),
    val albums: List<AlbumDto> = emptyList(),
    val playlists: List<PlaylistDto> = emptyList(),
    val folders: List<FolderDto> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class MusicViewModel @Inject constructor(
    private val repository: MusicRepository,
    private val player: PlayerConnection,
) : ViewModel() {

    private val _state = MutableStateFlow(MusicUiState())
    val state: StateFlow<MusicUiState> = _state.asStateFlow()

    init { loadAll() }

    fun selectTab(tab: MusicTab) {
        _state.update { it.copy(tab = tab) }
    }

    /** Recarrega a biblioteca exibindo o indicador de pull-to-refresh (sem spinner de tela cheia). */
    fun refresh() = loadAll(refreshing = true)

    /** Toca a faixa enfileirando a lista de contexto inteira a partir dela (comportamento Spotify). */
    fun play(track: TrackDto, context: List<TrackDto>) = player.play(track, context)

    private fun loadAll(refreshing: Boolean = false) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = !refreshing, isRefreshing = refreshing) }
            val tracks = repository.getAllTracks()
            val artists = repository.getArtists()
            val albums = repository.getAlbums()
            val playlists = repository.getPlaylists()
            val folders = repository.getFolders()
            _state.update {
                it.copy(
                    isLoading = false,
                    isRefreshing = false,
                    tracks = (tracks as? AppResult.Success)?.data ?: emptyList(),
                    artists = (artists as? AppResult.Success)?.data ?: emptyList(),
                    albums = (albums as? AppResult.Success)?.data ?: emptyList(),
                    playlists = (playlists as? AppResult.Success)?.data ?: emptyList(),
                    folders = (folders as? AppResult.Success)?.data ?: emptyList(),
                    error = (tracks as? AppResult.Error)?.message,
                )
            }
        }
    }
}
