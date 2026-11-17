package com.kuranas.android.feature.music.ui

import com.kuranas.android.feature.files.ui.FileStarErrorHost
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
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
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.PagedListFooter

@Composable
fun MusicAlbumScreen(
    albumKey: String,
    onNavigateBack: () -> Unit,
    onPlayTrack: () -> Unit,
) {
    val viewModel: MusicAlbumViewModel = hiltViewModel()
    LaunchedEffect(albumKey) { viewModel.load(albumKey) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val starredById by viewModel.stars.starredById.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.items.size, state.hasMore, viewModel::loadMore)

    FileStarErrorHost(viewModel.stars) {
        Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
            KNHeader(title = albumKey, leadingIcon = Icons.AutoMirrored.Filled.ArrowBack, onLeadingClick = onNavigateBack)
            if (state.isLoading) {
                LoadingView()
            } else {
                LazyColumn(state = listState, contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    items(state.items, key = { it.id }) { track ->
                        TrackListItem(
                            track = track,
                            isStarred = starredById[track.id.toString()] ?: track.starred,
                            onToggleStar = { viewModel.stars.toggle(track) },
                            onClick = {
                                viewModel.play(track)
                                onPlayTrack()
                            },
                        )
                    }
                    item {
                        PagedListFooter(isLoadingMore = state.isLoadingMore, loadMoreError = state.loadMoreError, onRetry = viewModel::retryLoadMore)
                    }
                }
            }
        }
    }
}
