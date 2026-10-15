package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.kuranas.android.R
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView

@Composable
fun VideoPlaylistScreen(
    playlistId: Int,
    onNavigateBack: () -> Unit,
    onPlayVideo: (String) -> Unit,
    viewModel: VideoPlaylistViewModel = hiltViewModel(),
    playerViewModel: VideoPlayerViewModel = hiltViewModel(),
) {
    LaunchedEffect(playlistId) { viewModel.load(playlistId) }
    val state by viewModel.state.collectAsState()
    val playlist = state.playlist
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.items.size, state.hasMore, viewModel::loadMore)

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = playlist?.name ?: stringResource(R.string.playlist_title),
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )
        val errorMessage = state.error
        when {
            state.isLoading -> LoadingView()
            errorMessage != null -> ErrorView(message = errorMessage, onRetry = viewModel::retry)
            state.items.isEmpty() -> EmptyView(message = stringResource(R.string.video_empty))
            else -> LazyColumn(
                state = listState,
                contentPadding = PaddingValues(bottom = 24.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(state.items, key = { it.video.id }) { item ->
                    VideoListItem(
                        video = item.video,
                        thumbnailUrl = playerViewModel.getThumbnailUrl(item.video.id),
                        onClick = { onPlayVideo(item.video.id) },
                    )
                }
            }
        }
    }
}
