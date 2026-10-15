package com.kuranas.android.feature.video.ui

import android.view.ViewGroup
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.SkipPrevious
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.exoplayer.DefaultRenderersFactory
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
import com.kuranas.android.R

@Composable
fun VideoPlayerScreen(
    videoId: String,
    onNavigateBack: () -> Unit,
    viewModel: VideoPlayerViewModel = hiltViewModel(),
) {
    val context = LocalContext.current
    val player = remember {
        // Capturas vêm de streams (H.264 High@L5.1); o decoder de hardware do
        // dispositivo pode recusar o nível (NO_EXCEEDS_CAPABILITIES). O fallback
        // permite cair para o decoder de software quando o hardware não inicializa.
        val renderersFactory = DefaultRenderersFactory(context)
            .setEnableDecoderFallback(true)
        ExoPlayer.Builder(context, renderersFactory).build()
    }
    var playbackError by remember { mutableStateOf<String?>(null) }

    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onPlayerError(error: PlaybackException) {
                playbackError = "${error.errorCodeName}: ${error.message ?: ""}"
            }

            override fun onPlayWhenReadyChanged(playWhenReady: Boolean, reason: Int) {
                if (!playWhenReady) viewModel.reportProgressNow(player.toPlayerProgress().copy(isPlaying = false))
            }
        }
        player.addListener(listener)
        viewModel.startProgressReporting { player.toPlayerProgress() }
        onDispose {
            viewModel.stopProgressReporting()
            viewModel.reportProgressNow(player.toPlayerProgress().copy(isPlaying = false))
            player.removeListener(listener)
            player.release()
        }
    }

    LaunchedEffect(videoId) {
        playbackError = null
        val start = viewModel.openVideo(videoId)
        player.setMediaItem(MediaItem.fromUri(start.streamUrl), start.startPositionMs)
        player.prepare()
        player.play()
    }

    val navigationAvailability by viewModel.navigationAvailability.collectAsState()
    val navigationError by viewModel.navigationError.collectAsState()

    LaunchedEffect(viewModel) {
        viewModel.mediaChanges.collect { start ->
            playbackError = null
            player.setMediaItem(MediaItem.fromUri(start.streamUrl), start.startPositionMs)
            player.prepare()
            player.play()
        }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        AndroidView(
            factory = { ctx ->
                PlayerView(ctx).apply {
                    this.player = player
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                }
            },
            modifier = Modifier.fillMaxSize(),
        )
        playbackError?.let { message ->
            Text(
                text = stringResource(R.string.video_playback_error, message),
                modifier = Modifier.align(Alignment.Center).padding(24.dp),
            )
        }
        navigationError?.let { message ->
            Text(
                text = message,
                modifier = Modifier.align(Alignment.BottomCenter).padding(24.dp),
            )
        }
        Row(
            modifier = Modifier.align(Alignment.TopStart).padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            IconButton(onClick = onNavigateBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.action_back))
            }
            IconButton(onClick = viewModel::goToPrevious, enabled = navigationAvailability.canGoPrevious) {
                Icon(Icons.Filled.SkipPrevious, contentDescription = stringResource(R.string.video_previous))
            }
            IconButton(onClick = viewModel::goToNext, enabled = navigationAvailability.canGoNext) {
                Icon(Icons.Filled.SkipNext, contentDescription = stringResource(R.string.video_next))
            }
        }
    }
}

private fun Player.toPlayerProgress() = PlayerProgress(
    positionMs = currentPosition,
    durationMs = duration,
    isPlaying = isPlaying,
)
