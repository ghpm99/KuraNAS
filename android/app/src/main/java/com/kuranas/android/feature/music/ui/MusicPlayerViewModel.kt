package com.kuranas.android.feature.music.ui

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.music.data.MusicRepository
import com.kuranas.android.feature.music.data.TrackDto
import com.kuranas.android.feature.music.playback.NowPlayingState
import com.kuranas.android.feature.music.playback.PlayerConnection
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * VM fino: o estado e o controle vivem no [PlayerConnection] (@Singleton), compartilhado
 * por mini-player, player completo e listas. Antes a lógica de MediaController ficava aqui
 * com escopo de tela, então o player abria sem saber o que tocar.
 *
 * Quando a tela é aberta a partir de Arquivos/Início/Busca, a rota traz um `trackId`: aqui
 * resolvemos a faixa e iniciamos a reprodução. Antes esses fluxos só navegavam para o
 * player sem nunca chamar play(), então a tela abria vazia e nada tocava.
 */
@HiltViewModel
class MusicPlayerViewModel @Inject constructor(
    private val player: PlayerConnection,
    private val repository: MusicRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    val state: StateFlow<NowPlayingState> = player.state

    init {
        val trackId = savedStateHandle.get<Int>("trackId") ?: -1
        // Só dispara se veio um id explícito e não é a faixa que já está tocando (evita
        // reiniciar ao reabrir o player pelo mini-player).
        if (trackId > 0 && player.state.value.currentTrack?.id != trackId) {
            viewModelScope.launch {
                val track = repository.getTrackById(trackId)
                player.play(track, listOf(track))
            }
        }
    }

    fun play(track: TrackDto, context: List<TrackDto>) = player.play(track, context)
    fun togglePlayPause() = player.togglePlayPause()
    fun next() = player.next()
    fun previous() = player.previous()
    fun seekTo(positionMs: Long) = player.seekTo(positionMs)
    fun toggleShuffle() = player.toggleShuffle()
    fun toggleRepeat() = player.cycleRepeat()
    fun skipToQueueItem(index: Int) = player.skipToQueueItem(index)
}
