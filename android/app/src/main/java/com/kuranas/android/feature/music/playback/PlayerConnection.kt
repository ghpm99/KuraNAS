package com.kuranas.android.feature.music.playback

import android.content.ComponentName
import android.content.Context
import android.net.Uri
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.Player
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import com.google.common.util.concurrent.ListenableFuture
import com.google.common.util.concurrent.MoreExecutors
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.ui.MusicPlaybackService
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.Dispatchers

/** Estado observável do player, compartilhado por mini-player, player completo e listas. */
data class NowPlayingState(
    val currentTrack: TrackDto? = null,
    val queue: List<TrackDto> = emptyList(),
    val currentIndex: Int = -1,
    val isPlaying: Boolean = false,
    val positionMs: Long = 0L,
    val durationMs: Long = 0L,
    val shuffle: Boolean = false,
    val repeatMode: Int = Player.REPEAT_MODE_OFF,
    val artworkUrl: String? = null,
) {
    val hasContent: Boolean get() = currentTrack != null

    /** Faixas que ainda vão tocar (depois da atual), preservando o índice real na fila. */
    val upNext: List<IndexedTrack>
        get() = if (currentIndex < 0) emptyList()
        else queue.drop(currentIndex + 1).mapIndexed { offset, track ->
            IndexedTrack(currentIndex + 1 + offset, track)
        }
}

data class IndexedTrack(val index: Int, val track: TrackDto)

/**
 * Dona única da [MediaController] (conecta uma vez, vive no escopo do app) e fonte de
 * verdade do estado de reprodução. Substitui a lógica antes presa ao MusicPlayerViewModel,
 * que ficava com escopo de tela e não era compartilhada — por isso o player abria vazio.
 */
@Singleton
class PlayerConnection @Inject constructor(
    @ApplicationContext private val context: Context,
    private val repository: MusicRepository,
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    private val _state = MutableStateFlow(NowPlayingState())
    val state: StateFlow<NowPlayingState> = _state.asStateFlow()

    private var controllerFuture: ListenableFuture<MediaController>? = null
    private var controller: MediaController? = null
    private var pendingAction: (() -> Unit)? = null

    init {
        connect()
        startPositionTicker()
    }

    private fun connect() {
        val token = SessionToken(context, ComponentName(context, MusicPlaybackService::class.java))
        controllerFuture = MediaController.Builder(context, token).buildAsync().also { future ->
            future.addListener({
                controller = future.get()
                controller?.addListener(listener)
                syncFromController()
                pendingAction?.invoke()
                pendingAction = null
            }, MoreExecutors.directExecutor())
        }
    }

    private val listener = object : Player.Listener {
        override fun onIsPlayingChanged(isPlaying: Boolean) {
            _state.update { it.copy(isPlaying = isPlaying) }
        }

        override fun onMediaItemTransition(mediaItem: MediaItem?, reason: Int) {
            updateCurrentFromController()
        }

        override fun onPlaybackStateChanged(playbackState: Int) {
            syncFromController()
        }

        override fun onShuffleModeEnabledChanged(shuffleModeEnabled: Boolean) {
            _state.update { it.copy(shuffle = shuffleModeEnabled) }
        }

        override fun onRepeatModeChanged(repeatMode: Int) {
            _state.update { it.copy(repeatMode = repeatMode) }
        }
    }

    // ---- Comandos ----

    /** Toca [tracks] como fila, começando em [startIndex] (Spotify: enfileira o contexto inteiro). */
    fun playContext(tracks: List<TrackDto>, startIndex: Int) {
        if (tracks.isEmpty()) return
        val index = startIndex.coerceIn(0, tracks.lastIndex)
        // Atualiza UI imediatamente (antes de resolver URLs / conectar).
        _state.update { it.copy(queue = tracks, currentIndex = index, currentTrack = tracks[index]) }
        scope.launch {
            val items = tracks.map { it.toMediaItem() }
            val artwork = repository.thumbnailUrl(tracks[index].id)
            _state.update { it.copy(artworkUrl = artwork) }
            runOrDefer {
                it.setMediaItems(items, index, 0L)
                it.prepare()
                it.play()
            }
        }
    }

    fun play(track: TrackDto, context: List<TrackDto>) {
        val list = context.ifEmpty { listOf(track) }
        playContext(list, list.indexOfFirst { it.id == track.id }.coerceAtLeast(0))
    }

    fun togglePlayPause() = controller?.let { if (it.isPlaying) it.pause() else it.play() }
    fun next() = controller?.seekToNextMediaItem() ?: Unit
    fun previous() = controller?.seekToPreviousMediaItem() ?: Unit

    fun seekTo(positionMs: Long) {
        controller?.seekTo(positionMs)
        _state.update { it.copy(positionMs = positionMs) }
    }

    fun skipToQueueItem(index: Int) {
        controller?.let {
            it.seekToDefaultPosition(index)
            it.play()
        }
    }

    fun toggleShuffle() {
        controller?.let { it.shuffleModeEnabled = !it.shuffleModeEnabled }
    }

    fun cycleRepeat() {
        val next = when (controller?.repeatMode ?: Player.REPEAT_MODE_OFF) {
            Player.REPEAT_MODE_OFF -> Player.REPEAT_MODE_ALL
            Player.REPEAT_MODE_ALL -> Player.REPEAT_MODE_ONE
            else -> Player.REPEAT_MODE_OFF
        }
        controller?.repeatMode = next
    }

    fun addToQueue(track: TrackDto) = enqueue(track, atEnd = true)
    fun playNext(track: TrackDto) = enqueue(track, atEnd = false)

    private fun enqueue(track: TrackDto, atEnd: Boolean) {
        scope.launch {
            val item = track.toMediaItem()
            runOrDefer { c ->
                val pos = if (atEnd) c.mediaItemCount else (c.currentMediaItemIndex + 1).coerceAtMost(c.mediaItemCount)
                c.addMediaItem(pos, item)
                rebuildQueueFromState(track, pos)
            }
        }
    }

    private fun rebuildQueueFromState(track: TrackDto, pos: Int) {
        _state.update {
            val newQueue = it.queue.toMutableList().apply { add(pos.coerceIn(0, size), track) }
            it.copy(queue = newQueue, currentIndex = controller?.currentMediaItemIndex ?: it.currentIndex)
        }
    }

    fun removeFromQueue(index: Int) {
        val c = controller ?: return
        if (index !in 0 until c.mediaItemCount) return
        c.removeMediaItem(index)
        _state.update {
            it.copy(
                queue = it.queue.toMutableList().apply { if (index in indices) removeAt(index) },
                currentIndex = c.currentMediaItemIndex,
            )
        }
    }

    // ---- Internos ----

    private inline fun runOrDefer(crossinline block: (MediaController) -> Unit) {
        val c = controller
        if (c != null) block(c) else pendingAction = { controller?.let { block(it) } }
    }

    private fun syncFromController() {
        val c = controller ?: return
        _state.update {
            it.copy(
                isPlaying = c.isPlaying,
                durationMs = c.duration.coerceAtLeast(0),
                shuffle = c.shuffleModeEnabled,
                repeatMode = c.repeatMode,
            )
        }
        if (c.mediaItemCount > 0) updateCurrentFromController()
    }

    private fun updateCurrentFromController() {
        val c = controller ?: return
        val index = c.currentMediaItemIndex
        val track = _state.value.queue.getOrNull(index)
        _state.update {
            it.copy(currentIndex = index, currentTrack = track ?: it.currentTrack, durationMs = c.duration.coerceAtLeast(0))
        }
        track?.let { t -> scope.launch { _state.update { s -> s.copy(artworkUrl = repository.thumbnailUrl(t.id)) } } }
    }

    private fun startPositionTicker() {
        scope.launch {
            while (true) {
                controller?.let { c ->
                    if (c.isPlaying) {
                        _state.update {
                            it.copy(positionMs = c.currentPosition.coerceAtLeast(0), durationMs = c.duration.coerceAtLeast(0))
                        }
                    }
                }
                delay(500)
            }
        }
    }

    private suspend fun TrackDto.toMediaItem(): MediaItem {
        val metadata = MediaMetadata.Builder()
            .setTitle(title)
            .setArtist(artist ?: "")
            .setAlbumTitle(album ?: "")
            .setArtworkUri(runCatching { Uri.parse(repository.thumbnailUrl(id)) }.getOrNull())
            .build()
        return MediaItem.Builder()
            .setMediaId(id.toString())
            .setUri(repository.streamUrl(id))
            .setMediaMetadata(metadata)
            .build()
    }
}
