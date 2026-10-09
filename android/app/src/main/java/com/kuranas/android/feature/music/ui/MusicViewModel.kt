package com.kuranas.android.feature.music.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
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
import javax.inject.Inject

enum class MusicTab { TRACKS, ARTISTS, ALBUMS, PLAYLISTS, FOLDERS }

@HiltViewModel
class MusicViewModel @Inject constructor(
    private val repository: MusicRepository,
    private val player: PlayerConnection,
) : ViewModel() {

    private val _tab = MutableStateFlow(MusicTab.TRACKS)
    val tab: StateFlow<MusicTab> = _tab.asStateFlow()

    private val tracksLoader = PagedListLoader<TrackDto>(viewModelScope) { it.id }
    private val artistsLoader = PagedListLoader<ArtistDto>(viewModelScope) { it.key }
    private val albumsLoader = PagedListLoader<AlbumDto>(viewModelScope) { it.key }
    private val playlistsLoader = PagedListLoader<PlaylistDto>(viewModelScope) { it.id }
    private val foldersLoader = PagedListLoader<FolderDto>(viewModelScope) { it.folder }

    val tracks: StateFlow<PagedListState<TrackDto>> = tracksLoader.state
    val artists: StateFlow<PagedListState<ArtistDto>> = artistsLoader.state
    val albums: StateFlow<PagedListState<AlbumDto>> = albumsLoader.state
    val playlists: StateFlow<PagedListState<PlaylistDto>> = playlistsLoader.state
    val folders: StateFlow<PagedListState<FolderDto>> = foldersLoader.state

    init {
        tracksLoader.load(repository::getAllTracks)
        artistsLoader.load(repository::getArtists)
        albumsLoader.load(repository::getAlbums)
        playlistsLoader.load(repository::getPlaylists)
        foldersLoader.load(repository::getFolders)
    }

    fun selectTab(tab: MusicTab) {
        _tab.update { tab }
    }

    fun refresh() {
        tracksLoader.refresh()
        artistsLoader.refresh()
        albumsLoader.refresh()
        playlistsLoader.refresh()
        foldersLoader.refresh()
    }

    fun loadMoreTracks() = tracksLoader.loadMore()
    fun loadMoreArtists() = artistsLoader.loadMore()
    fun loadMoreAlbums() = albumsLoader.loadMore()
    fun loadMorePlaylists() = playlistsLoader.loadMore()
    fun loadMoreFolders() = foldersLoader.loadMore()

    fun play(track: TrackDto, context: List<TrackDto>) = player.play(track, context)
}
