package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistSection

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VideoScreen(
    onPlayVideo: (String) -> Unit,
    onOpenPlaylist: (Int) -> Unit,
    viewModel: VideoViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = stringResource(R.string.nav_videos))
        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
        when {
            state.isLoading -> LoadingView()
            state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
            state.catalog == null -> EmptyView(stringResource(R.string.video_empty))
            else -> {
                val catalog = state.catalog!!
                LazyColumn(modifier = Modifier.fillMaxSize(), contentPadding = PaddingValues(bottom = 24.dp)) {
                    if (catalog.recentVideos.isNotEmpty()) {
                        item {
                            Text(stringResource(R.string.video_recent), style = MaterialTheme.typography.titleMedium)
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
                    items(state.playlistSections, key = { it.section }) { sectionState ->
                        PlaylistSection(
                            sectionState = sectionState,
                            thumbnailUrl = viewModel::thumbnailUrl,
                            onOpenPlaylist = onOpenPlaylist,
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
private fun PlaylistSection(
    sectionState: VideoPlaylistSectionState,
    thumbnailUrl: (String) -> String,
    onOpenPlaylist: (Int) -> Unit,
) {
    Column(modifier = Modifier.padding(bottom = 16.dp)) {
        Text(stringResource(sectionState.section.titleRes()), style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(8.dp))
        LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            items(sectionState.playlists, key = { it.id }) { playlist ->
                PlaylistCard(
                    playlist = playlist,
                    coverUrl = playlist.coverVideoId?.let { thumbnailUrl(it.toString()) },
                    onClick = { onOpenPlaylist(playlist.id) },
                )
            }
        }
    }
}

@Composable
private fun PlaylistCard(playlist: VideoPlaylistDto, coverUrl: String?, onClick: () -> Unit) {
    Column(
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(8.dp)
            .size(width = 160.dp, height = 140.dp),
    ) {
        AsyncImage(
            model = coverUrl,
            contentDescription = playlist.name,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .clip(RoundedCornerShape(8.dp)),
        )
        Text(playlist.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.padding(top = 4.dp))
        Text(pluralStringResource(R.plurals.video_count, playlist.count, playlist.count), style = MaterialTheme.typography.bodySmall)
    }
}

private fun VideoPlaylistSection.titleRes(): Int = when (this) {
    VideoPlaylistSection.Series -> R.string.video_section_series
    VideoPlaylistSection.Movies -> R.string.video_section_movies
    VideoPlaylistSection.Personal -> R.string.video_section_personal
    VideoPlaylistSection.Clips -> R.string.video_section_clips
    VideoPlaylistSection.Folders -> R.string.video_section_folders
}
