package com.kuranas.android.feature.documents.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.WindowInsetsSides
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.only
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.ReconnectionRefresh
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.PagedListFooter
import com.kuranas.android.core.ui.components.formatBytes
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.documents.data.DocumentDto

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentsScreen(onNavigateBack: () -> Unit, viewModel: DocumentsViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.documents.size, state.hasMore, viewModel::loadMore)

    LaunchedEffect(Unit) { viewModel.startIfNeeded() }
    ReconnectionRefresh(viewModel::refresh)

    Scaffold(
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(
                title = stringResource(R.string.nav_documents),
                leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
                onLeadingClick = onNavigateBack,
            )
            state.totalDocuments?.let { total ->
                Text(
                    pluralStringResource(R.plurals.documents_total, total, total),
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(horizontal = 4.dp, vertical = 4.dp),
                )
            }
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                val errorMessage = state.error
                when {
                    state.isLoading -> LoadingView()
                    errorMessage != null -> ErrorView(message = errorMessage, onRetry = viewModel::reload)
                    state.documents.isEmpty() -> EmptyView(message = stringResource(R.string.documents_empty_state))
                    else -> LazyColumn(
                        state = listState,
                        contentPadding = PaddingValues(top = 8.dp, bottom = 80.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        items(state.documents, key = { it.id }) { document -> DocumentRow(document) }
                        item {
                            PagedListFooter(
                                isLoadingMore = state.isLoadingMore,
                                loadMoreError = state.loadMoreError,
                                onRetry = viewModel::retryLoadMore,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun DocumentRow(document: DocumentDto) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(
            imageVector = documentFormatIcon(document.format),
            contentDescription = null,
            modifier = Modifier.size(24.dp),
            tint = MaterialTheme.colorScheme.secondary,
        )
        Column(modifier = Modifier.weight(1f)) {
            Text(document.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                document.parentPath,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                stringResource(R.string.documents_size_and_date, formatBytes(document.size), formatDocumentDate(document.modifiedAt)),
                style = MaterialTheme.typography.bodySmall,
            )
        }
        if (document.starred) {
            Icon(
                imageVector = Icons.Default.Star,
                contentDescription = stringResource(R.string.documents_starred),
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(20.dp),
            )
        }
    }
}
