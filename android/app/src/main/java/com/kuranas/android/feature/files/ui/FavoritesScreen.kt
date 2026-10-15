package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.WindowInsetsSides
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.only
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
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
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.feature.files.data.StarredKind
import com.kuranas.android.feature.files.data.StarredSort

@Composable
fun FavoritesScreen(
    onOpenFolder: (String, String) -> Unit,
    onOpenImage: (String) -> Unit,
    onOpenVideo: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onOpenFile: (String, String) -> Unit,
    viewModel: FavoritesViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.favorites.size, state.hasMore, viewModel::loadMore)

    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(state.actionErrorMessage) {
        val message = state.actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissActionError()
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(title = stringResource(R.string.nav_favorites))
            KindChips(selected = state.kind, onSelect = viewModel::selectKind)
            SortChips(selected = state.sort, onSelect = viewModel::selectSort)
            val errorMessage = state.error
            when {
                state.isLoading -> LoadingView()
                errorMessage != null -> ErrorView(message = errorMessage, onRetry = viewModel::retry)
                state.favorites.isEmpty() -> EmptyView(message = stringResource(R.string.favorites_empty))
                else -> LazyColumn(
                    state = listState,
                    contentPadding = PaddingValues(top = 8.dp, bottom = 80.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    items(state.favorites, key = { it.id }) { file ->
                        FileListItem(
                            file = file,
                            onClick = {
                                if (file.isDir) onOpenFolder(file.id, file.name)
                                else openFileByType(file, onOpenImage, onOpenVideo, onPlayAudio, onOpenFile)
                            },
                            onStar = { viewModel.unstar(file) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun KindChips(selected: StarredKind, onSelect: (StarredKind) -> Unit) {
    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        items(StarredKind.entries) { kind ->
            FilterChip(
                selected = kind == selected,
                onClick = { onSelect(kind) },
                label = { Text(stringResource(kind.labelRes())) },
            )
        }
    }
}

@Composable
private fun SortChips(selected: StarredSort, onSelect: (StarredSort) -> Unit) {
    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.padding(top = 4.dp)) {
        items(StarredSort.entries) { sort ->
            FilterChip(
                selected = sort == selected,
                onClick = { onSelect(sort) },
                label = { Text(stringResource(sort.labelRes())) },
            )
        }
    }
}

private fun StarredKind.labelRes() = when (this) {
    StarredKind.ALL -> R.string.favorites_kind_all
    StarredKind.FOLDERS -> R.string.favorites_kind_folders
    StarredKind.FILES -> R.string.favorites_kind_files
    StarredKind.MEDIA -> R.string.favorites_kind_media
}

private fun StarredSort.labelRes() = when (this) {
    StarredSort.NAME -> R.string.favorites_sort_name
    StarredSort.STARRED_AT -> R.string.favorites_sort_starred_at
}
