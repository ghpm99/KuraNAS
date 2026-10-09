package com.kuranas.android.feature.images.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView

private const val PREFETCH_DISTANCE_ITEMS = 24

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ImagesScreen(
    onOpenImage: (String) -> Unit,
    viewModel: ImagesViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    val gridState = rememberLazyGridState()
    val isNearEnd by remember {
        derivedStateOf {
            val lastVisibleIndex = gridState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisibleIndex >= gridState.layoutInfo.totalItemsCount - PREFETCH_DISTANCE_ITEMS
        }
    }
    LaunchedEffect(isNearEnd, state.images.size, state.hasMore) {
        if (isNearEnd) viewModel.loadMore()
    }

    var firstResume by rememberSaveable { mutableStateOf(true) }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (firstResume) firstResume = false else viewModel.refresh()
    }

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = stringResource(R.string.nav_images))
        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
            when {
                state.isLoading -> LoadingView()
                state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
                state.images.isEmpty() -> EmptyView(stringResource(R.string.images_empty))
                else -> LazyVerticalGrid(
                    state = gridState,
                    columns = GridCells.Adaptive(minSize = 120.dp),
                    contentPadding = PaddingValues(bottom = 24.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    items(state.images, key = { it.id }) { image ->
                        AsyncImage(
                            model = viewModel.thumbnailUrl(image.id),
                            contentDescription = image.name,
                            contentScale = ContentScale.Crop,
                            modifier = Modifier
                                .aspectRatio(1f)
                                .clip(RoundedCornerShape(8.dp))
                                .clickable { onOpenImage(image.id) },
                        )
                    }
                    if (state.isLoadingMore) {
                        item(span = { GridItemSpan(maxLineSpan) }) {
                            CircularProgressIndicator(
                                modifier = Modifier.fillMaxWidth().padding(16.dp).wrapContentSize(Alignment.Center),
                            )
                        }
                    }
                }
            }
        }
    }
}
