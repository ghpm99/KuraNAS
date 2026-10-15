package com.kuranas.android.feature.video.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass

@Composable
fun VideoLibraryScreen(
    onNavigateBack: () -> Unit,
    onPlayVideo: (String) -> Unit,
    viewModel: VideoLibraryViewModel = hiltViewModel(),
    playerViewModel: VideoPlayerViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.videos.size, state.hasMore, viewModel::loadMore)

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = stringResource(R.string.video_library_title),
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )
        TextField(
            value = state.searchQuery,
            onValueChange = viewModel::onSearchQueryChange,
            placeholder = { Text(stringResource(R.string.video_library_search_placeholder)) },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            singleLine = true,
            colors = TextFieldDefaults.colors(
                focusedContainerColor = Color.Transparent,
                unfocusedContainerColor = Color.Transparent,
            ),
            modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light, radius = 16.dp),
        )
        Spacer(Modifier.height(8.dp))
        val errorMessage = state.error
        when {
            state.isLoading -> LoadingView()
            errorMessage != null -> ErrorView(message = errorMessage, onRetry = viewModel::retry)
            state.videos.isEmpty() -> EmptyView(message = stringResource(R.string.video_empty))
            else -> LazyColumn(
                state = listState,
                contentPadding = PaddingValues(bottom = 24.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(state.videos, key = { it.id }) { video ->
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
