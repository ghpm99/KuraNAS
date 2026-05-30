package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.PlayCircle
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto

@Composable
fun VideoScreen(
    onPlayVideo: (String) -> Unit,
    onOpenPlaylist: (Int) -> Unit,
    viewModel: VideoViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = "Vídeos")
        when {
            state.isLoading -> LoadingView()
            state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
            state.catalog == null -> EmptyView("Nenhum vídeo encontrado")
            else -> {
                val catalog = state.catalog!!
                LazyColumn(contentPadding = PaddingValues(bottom = 24.dp)) {
                    if (catalog.recentVideos.isNotEmpty()) {
                        item {
                            Text("Recentes", style = MaterialTheme.typography.titleMedium)
                            Spacer(Modifier.height(8.dp))
                            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                items(catalog.recentVideos) { video ->
                                    VideoCard(
                                        video = video,
                                        thumbnailUrl = viewModel.thumbnailUrl(video.id),
                                        onClick = { onPlayVideo(video.id) },
                                    )
                                }
                            }
                            Spacer(Modifier.height(16.dp))
                        }
                    }
                    if (catalog.playlists.isNotEmpty()) {
                        item { Text("Playlists", style = MaterialTheme.typography.titleMedium) }
                        items(catalog.playlists, key = { it.id }) { playlist ->
                            PlaylistRow(
                                playlist = playlist,
                                onOpen = { onOpenPlaylist(playlist.id) },
                                onPlayFirst = {
                                    playlist.videos.firstOrNull()?.let { onPlayVideo(it.id) }
                                },
                                thumbnailUrl = { id -> viewModel.thumbnailUrl(id) },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun VideoCard(video: VideoItemDto, thumbnailUrl: String, onClick: () -> Unit) {
    Column(
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(8.dp)
            .size(width = 160.dp, height = 120.dp),
    ) {
        AsyncImage(
            model = thumbnailUrl,
            contentDescription = video.name,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .clip(RoundedCornerShape(8.dp)),
        )
        Text(video.name, style = MaterialTheme.typography.bodySmall, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.padding(top = 4.dp))
    }
}

@Composable
private fun PlaylistRow(
    playlist: VideoPlaylistDto,
    onOpen: () -> Unit,
    onPlayFirst: () -> Unit,
    thumbnailUrl: (String) -> String,
) {
    Column(modifier = Modifier.padding(vertical = 8.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth().clickable(onClick = onOpen),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(playlist.name, style = MaterialTheme.typography.titleMedium)
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("${playlist.count} vídeos", style = MaterialTheme.typography.bodySmall)
                Icon(Icons.Default.ChevronRight, contentDescription = null, modifier = Modifier.size(20.dp))
            }
        }
        Spacer(Modifier.height(8.dp))
        LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            items(playlist.videos.take(5)) { video ->
                AsyncImage(
                    model = thumbnailUrl(video.id),
                    contentDescription = video.name,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .size(width = 120.dp, height = 80.dp)
                        .clip(RoundedCornerShape(8.dp))
                        .clickable { onPlayFirst() },
                )
            }
        }
    }
}
