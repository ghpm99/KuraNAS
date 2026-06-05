package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.video.data.VideoItemDto

@Composable
fun VideoPlaylistScreen(
    playlistId: Int,
    onNavigateBack: () -> Unit,
    onPlayVideo: (String) -> Unit,
    viewModel: VideoPlaylistViewModel = hiltViewModel(),
    playerViewModel: VideoPlayerViewModel = hiltViewModel(),
) {
    LaunchedEffect(playlistId) { viewModel.load(playlistId) }
    val playlist = viewModel.playlist

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = playlist?.name ?: stringResource(R.string.playlist_title),
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )
        if (playlist == null) {
            LoadingView()
        } else {
            LazyColumn(
                contentPadding = PaddingValues(bottom = 24.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(playlist.videos, key = { it.id }) { video ->
                    VideoListItem(
                        video = video,
                        thumbnailUrl = playerViewModel.getThumbnailUrl(video.id),
                        onClick = { onPlayVideo(video.id) },
                    )
                }
            }
        }
    }
}

@Composable
private fun VideoListItem(video: VideoItemDto, thumbnailUrl: String, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onClick)
            .padding(8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        AsyncImage(
            model = thumbnailUrl,
            contentDescription = video.name,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .size(width = 80.dp, height = 52.dp)
                .clip(RoundedCornerShape(8.dp)),
        )
        Text(video.name, style = MaterialTheme.typography.bodyMedium, modifier = Modifier.weight(1f), maxLines = 2)
        Icon(Icons.Default.PlayArrow, contentDescription = stringResource(R.string.cd_play), tint = MaterialTheme.colorScheme.primary)
    }
}
