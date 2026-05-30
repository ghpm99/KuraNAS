package com.kuranas.android.feature.music.ui

import android.content.ComponentName
import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import com.google.common.util.concurrent.ListenableFuture
import com.google.common.util.concurrent.MoreExecutors
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.PlayerStateDto
import com.kuranas.android.feature.music.data.TrackDto
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class PlayerUiState(
    val currentTrack: TrackDto? = null,
    val isPlaying: Boolean = false,
    val position: Long = 0L,
    val duration: Long = 0L,
    val shuffle: Boolean = false,
    val repeatMode: Int = Player.REPEAT_MODE_OFF,
    val queue: List<TrackDto> = emptyList(),
)

@HiltViewModel
class MusicPlayerViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val repository: MusicRepository,
) : ViewModel() {

    private val _state = MutableStateFlow(PlayerUiState())
    val state: StateFlow<PlayerUiState> = _state.asStateFlow()

    private var controllerFuture: ListenableFuture<MediaController>? = null
    private var controller: MediaController? = null

    init {
        connectToService()
        loadPlayerState()
    }

    private fun connectToService() {
        val token = SessionToken(context, ComponentName(context, MusicPlaybackService::class.java))
        controllerFuture = MediaController.Builder(context, token).buildAsync()
        controllerFuture?.addListener({
            controller = controllerFuture?.get()
            controller?.addListener(playerListener)
        }, MoreExecutors.directExecutor())
    }

    private val playerListener = object : Player.Listener {
        override fun onIsPlayingChanged(isPlaying: Boolean) {
            _state.update { it.copy(isPlaying = isPlaying) }
        }
        override fun onPlaybackStateChanged(state: Int) {
            controller?.let { c ->
                _state.update { it.copy(duration = c.duration.coerceAtLeast(0)) }
            }
        }
    }

    fun playTrack(track: TrackDto, queue: List<TrackDto> = emptyList()) {
        viewModelScope.launch {
            val url = repository.streamUrl(track.id)
            val items = if (queue.isEmpty()) listOf(track) else queue
            val mediaItems = items.map { MediaItem.fromUri(repository.streamUrl(it.id)) }
            val startIndex = items.indexOfFirst { it.id == track.id }.coerceAtLeast(0)
            controller?.apply {
                setMediaItems(mediaItems, startIndex, 0L)
                prepare()
                play()
            }
            _state.update { it.copy(currentTrack = track, queue = items, isPlaying = true) }
        }
    }

    fun togglePlayPause() {
        controller?.let { if (it.isPlaying) it.pause() else it.play() }
    }

    fun next() { controller?.seekToNextMediaItem() }
    fun previous() { controller?.seekToPreviousMediaItem() }

    fun seekTo(position: Long) {
        controller?.seekTo(position)
        _state.update { it.copy(position = position) }
    }

    fun toggleShuffle() {
        val newShuffle = !_state.value.shuffle
        controller?.shuffleModeEnabled = newShuffle
        _state.update { it.copy(shuffle = newShuffle) }
    }

    fun toggleRepeat() {
        val newMode = when (_state.value.repeatMode) {
            Player.REPEAT_MODE_OFF -> Player.REPEAT_MODE_ALL
            Player.REPEAT_MODE_ALL -> Player.REPEAT_MODE_ONE
            else -> Player.REPEAT_MODE_OFF
        }
        controller?.repeatMode = newMode
        _state.update { it.copy(repeatMode = newMode) }
    }

    fun updatePosition() {
        controller?.let { c -> _state.update { it.copy(position = c.currentPosition.coerceAtLeast(0)) } }
    }

    private fun loadPlayerState() {
        viewModelScope.launch {
            repository.getPlayerState()
        }
    }

    override fun onCleared() {
        controller?.removeListener(playerListener)
        MediaController.releaseFuture(controllerFuture ?: return)
        super.onCleared()
    }
}
