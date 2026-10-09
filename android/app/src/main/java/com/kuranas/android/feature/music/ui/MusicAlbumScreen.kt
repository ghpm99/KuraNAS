package com.kuranas.android.feature.music.ui

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

@Composable
fun MusicAlbumScreen(
    albumKey: String,
    onNavigateBack: () -> Unit,
    onPlayTrack: () -> Unit,
) {
    val viewModel: MusicAlbumViewModel = hiltViewModel()
    LaunchedEffect(albumKey) { viewModel.load(albumKey) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.items.size, state.hasMore, viewModel::loadMore)

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = albumKey, leadingIcon = Icons.AutoMirrored.Filled.ArrowBack, onLeadingClick = onNavigateBack)
        if (state.isLoading) {
            LoadingView()
        } else {
            LazyColumn(state = listState, contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                items(state.items, key = { it.id }) { track ->
                    TrackListItem(track = track, onClick = {
                        viewModel.play(track)
                        onPlayTrack()
                    })
                }
            }
        }
    }
}
