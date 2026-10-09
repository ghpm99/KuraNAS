package com.kuranas.android.feature.music.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.PlayerConnection
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

@HiltViewModel
class MusicArtistViewModel @Inject constructor(
    val repository: MusicRepository,
    private val player: PlayerConnection,
) : ViewModel() {
    private val tracksLoader = PagedListLoader<TrackDto>(viewModelScope) { it.id }
    val state: StateFlow<PagedListState<TrackDto>> = tracksLoader.state

    fun load(key: String) = tracksLoader.load { page -> repository.getTracksByArtist(key, page) }

    fun loadMore() = tracksLoader.loadMore()

    fun play(track: TrackDto) = player.play(track, state.value.items)
}
