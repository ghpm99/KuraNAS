package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.VideoLibrary
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
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
import com.kuranas.android.feature.files.ui.FileStarButton
import com.kuranas.android.feature.files.ui.FileStarErrorHost
import com.kuranas.android.feature.video.data.VideoCatalogItemDto
import com.kuranas.android.feature.video.data.VideoHomeCatalogDto
import com.kuranas.android.feature.video.data.VideoItemDto
import com.kuranas.android.feature.video.data.VideoPlaylistDto
import com.kuranas.android.feature.video.data.VideoPlaylistSection

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VideoScreen(
    onPlayVideo: (String) -> Unit,
    onOpenPlaylist: (Int) -> Unit,
    onOpenLibrary: () -> Unit,
    onOpenSearch: () -> Unit,
    viewModel: VideoViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val starredById by viewModel.stars.starredById.collectAsStateWithLifecycle()
    val catalog = state.catalog
    val errorMessage = state.error

    FileStarErrorHost(viewModel.stars) {
        Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
            KNHeader(title = stringResource(R.string.nav_videos), onSearchClick = onOpenSearch)
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                when {
                    state.isLoading -> LoadingView()
                    errorMessage != null -> ErrorView(errorMessage, onRetry = viewModel::load)
                    catalog == null -> EmptyView(stringResource(R.string.video_empty))
                    else -> VideoCatalog(
                        catalog = catalog,
                        playlistSections = state.playlistSections,
                        starredById = starredById,
                        thumbnailUrl = viewModel::thumbnailUrl,
                        onToggleStar = viewModel.stars::toggle,
                        onPlayVideo = onPlayVideo,
                        onOpenPlaylist = onOpenPlaylist,
                        onOpenLibrary = onOpenLibrary,
                    )
                }
            }
        }
    }
}

@Composable
private fun VideoCatalog(
    catalog: VideoHomeCatalogDto,
    playlistSections: List<VideoPlaylistSectionState>,
    starredById: Map<String, Boolean>,
    thumbnailUrl: (String) -> String,
    onToggleStar: (VideoItemDto) -> Unit,
    onPlayVideo: (String) -> Unit,
    onOpenPlaylist: (Int) -> Unit,
    onOpenLibrary: () -> Unit,
) {
    LazyColumn(modifier = Modifier.fillMaxSize(), contentPadding = PaddingValues(bottom = 24.dp)) {
        item {
            AllVideosEntry(onClick = onOpenLibrary)
            Spacer(Modifier.height(16.dp))
        }
        if (catalog.continueWatchingItems.isNotEmpty()) {
            item {
                VideoCardRow(
                    title = stringResource(R.string.video_continue_watching),
                    catalogItems = catalog.continueWatchingItems,
                    starredById = starredById,
                    thumbnailUrl = thumbnailUrl,
                    onToggleStar = onToggleStar,
                    onPlayVideo = onPlayVideo,
                )
            }
        }
        if (catalog.recentVideos.isNotEmpty()) {
            item {
                VideoCardRow(
                    title = stringResource(R.string.video_recent),
                    catalogItems = catalog.recentVideos.map { video -> VideoCatalogItemDto(video = video) },
                    starredById = starredById,
                    thumbnailUrl = thumbnailUrl,
                    onToggleStar = onToggleStar,
                    onPlayVideo = onPlayVideo,
                    showsProgress = false,
                )
            }
        }
        items(playlistSections, key = { it.section }) { sectionState ->
            PlaylistSection(
                sectionState = sectionState,
                thumbnailUrl = thumbnailUrl,
                onOpenPlaylist = onOpenPlaylist,
            )
        }
    }
}

@Composable
private fun VideoCardRow(
    title: String,
    catalogItems: List<VideoCatalogItemDto>,
    starredById: Map<String, Boolean>,
    thumbnailUrl: (String) -> String,
    onToggleStar: (VideoItemDto) -> Unit,
    onPlayVideo: (String) -> Unit,
    showsProgress: Boolean = true,
) {
    Text(title, style = MaterialTheme.typography.titleMedium)
    Spacer(Modifier.height(8.dp))
    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        items(catalogItems, key = { it.video.id }) { catalogItem ->
            val video = catalogItem.video
            VideoCard(
                video = video,
                thumbnailUrl = thumbnailUrl(video.id),
                progressPercent = catalogItem.progressPct.takeIf { showsProgress },
                isStarred = starredById[video.id] ?: video.starred,
                onToggleStar = { onToggleStar(video) },
                onClick = { onPlayVideo(video.id) },
            )
        }
    }
    Spacer(Modifier.height(16.dp))
}

@Composable
private fun AllVideosEntry(onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(Icons.Default.VideoLibrary, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
        Text(stringResource(R.string.video_library_entry), style = MaterialTheme.typography.titleMedium, modifier = Modifier.weight(1f))
        Icon(Icons.AutoMirrored.Filled.KeyboardArrowRight, contentDescription = null)
    }
}

@Composable
private fun VideoCard(
    video: VideoItemDto,
    thumbnailUrl: String,
    isStarred: Boolean,
    onToggleStar: () -> Unit,
    onClick: () -> Unit,
    progressPercent: Double? = null,
) {
    Column(
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(8.dp)
            .size(width = 160.dp, height = if (progressPercent == null) 120.dp else 132.dp),
    ) {
        Box(modifier = Modifier.fillMaxWidth().weight(1f)) {
            AsyncImage(
                model = thumbnailUrl,
                contentDescription = video.name,
                contentScale = ContentScale.Crop,
                modifier = Modifier
                    .fillMaxSize()
                    .clip(RoundedCornerShape(8.dp)),
            )
            FileStarButton(
                isStarred = isStarred,
                onToggleStar = onToggleStar,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(4.dp)
                    .size(32.dp)
                    .background(MaterialTheme.colorScheme.surface.copy(alpha = 0.6f), CircleShape),
            )
        }
        Text(video.name, style = MaterialTheme.typography.bodySmall, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.padding(top = 4.dp))
        if (progressPercent != null) {
            LinearProgressIndicator(
                progress = { (progressPercent / 100.0).coerceIn(0.0, 1.0).toFloat() },
                modifier = Modifier.fillMaxWidth().padding(top = 4.dp),
            )
        }
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
