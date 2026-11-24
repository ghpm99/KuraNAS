package com.kuranas.android.feature.documents.ui

import android.content.Context
import android.widget.Toast
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.FilterChip
import androidx.compose.material3.IconButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
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
import com.kuranas.android.feature.files.data.download.toMessage
import com.kuranas.android.feature.files.ui.ExternalDelivery
import com.kuranas.android.feature.files.ui.buildExternalFileIntent

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentsScreen(onNavigateBack: () -> Unit, onOpenInViewer: (Int, String) -> Unit, viewModel: DocumentsViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.documents.size, state.hasMore, viewModel::loadMore)

    LaunchedEffect(Unit) { viewModel.startIfNeeded() }
    ReconnectionRefresh(viewModel::refresh)

    val context = LocalContext.current
    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is DocumentsEvent.OpenInViewer -> onOpenInViewer(event.documentId, event.documentName)
                is DocumentsEvent.OpenExternally -> context.deliverExternally(event.delivery)
                DocumentsEvent.DownloadQueued ->
                    Toast.makeText(context, R.string.documents_download_started, Toast.LENGTH_SHORT).show()
            }
        }
    }

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
            DocumentSearchField(text = state.searchText, onTextChange = viewModel::changeSearchText, onClear = viewModel::clearSearch)
            DocumentFormatChips(selected = state.formatFilter, onSelect = viewModel::selectFormat)
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
                    state.documents.isEmpty() -> EmptyView(
                        message = stringResource(if (state.isSearching) R.string.documents_search_empty else R.string.documents_empty_state),
                    )
                    else -> LazyColumn(
                        state = listState,
                        contentPadding = PaddingValues(top = 8.dp, bottom = 80.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        items(state.documents, key = { it.id }) { document -> DocumentRow(
                                document = document,
                                isPreparing = state.preparingDocumentId == document.id,
                                onOpen = { viewModel.open(document) },
                                onDownload = { viewModel.download(document) },
                            )
                        }
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
private fun DocumentSearchField(text: String, onTextChange: (String) -> Unit, onClear: () -> Unit) {
    OutlinedTextField(
        value = text,
        onValueChange = onTextChange,
        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
        singleLine = true,
        placeholder = { Text(stringResource(R.string.documents_search_placeholder)) },
        leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
        trailingIcon = {
            if (text.isNotEmpty()) {
                IconButton(onClick = onClear) {
                    Icon(Icons.Default.Close, contentDescription = stringResource(R.string.documents_clear_search))
                }
            }
        },
    )
}

@Composable
private fun DocumentFormatChips(selected: DocumentFormatFilter, onSelect: (DocumentFormatFilter) -> Unit) {
    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        items(DocumentFormatFilter.entries) { filter ->
            FilterChip(
                selected = filter == selected,
                onClick = { onSelect(filter) },
                label = { Text(stringResource(filter.labelRes())) },
            )
        }
    }
}

private fun DocumentFormatFilter.labelRes() = when (this) {
    DocumentFormatFilter.ALL -> R.string.documents_filter_all
    DocumentFormatFilter.PDF -> R.string.documents_filter_pdf
    DocumentFormatFilter.TEXT -> R.string.documents_filter_text
    DocumentFormatFilter.SPREADSHEET -> R.string.documents_filter_spreadsheet
    DocumentFormatFilter.PRESENTATION -> R.string.documents_filter_presentation
    DocumentFormatFilter.OTHER -> R.string.documents_filter_other
}

private fun Context.deliverExternally(delivery: ExternalDelivery) {
    when (delivery) {
        is ExternalDelivery.Ready -> startActivity(buildExternalFileIntent(this, delivery))
        is ExternalDelivery.Failed -> Toast.makeText(this, delivery.failure.toMessage(this), Toast.LENGTH_LONG).show()
    }
}

@Composable
private fun DocumentRow(document: DocumentListItem, isPreparing: Boolean, onOpen: () -> Unit, onDownload: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onOpen)
            .padding(start = 12.dp, top = 10.dp, bottom = 10.dp, end = 4.dp),
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
            document.snippet?.let { snippet ->
                Text(snippet, style = MaterialTheme.typography.bodySmall, maxLines = 3, overflow = TextOverflow.Ellipsis)
            }
            Text(
                document.parentPath,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                stringResource(R.string.documents_size_and_date, formatBytes(document.size), formatDocumentDate(document.timestamp)),
                style = MaterialTheme.typography.bodySmall,
            )
            document.pageNumber?.let { pageNumber ->
                Text(
                    stringResource(R.string.documents_page_number, pageNumber),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.primary,
                )
            }
        }
        if (isPreparing) {
            CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
        }
        if (document.isStarred) {
            Icon(
                imageVector = Icons.Default.Star,
                contentDescription = stringResource(R.string.documents_starred),
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(20.dp),
            )
        }
        DocumentActionsMenu(onOpen = onOpen, onDownload = onDownload)
    }
}

@Composable
private fun DocumentActionsMenu(onOpen: () -> Unit, onDownload: () -> Unit) {
    var isExpanded by remember { mutableStateOf(false) }
    Box {
        IconButton(onClick = { isExpanded = true }) {
            Icon(Icons.Default.MoreVert, contentDescription = stringResource(R.string.documents_more_actions))
        }
        DropdownMenu(expanded = isExpanded, onDismissRequest = { isExpanded = false }) {
            DropdownMenuItem(
                text = { Text(stringResource(R.string.documents_action_open)) },
                onClick = {
                    isExpanded = false
                    onOpen()
                },
            )
            DropdownMenuItem(
                text = { Text(stringResource(R.string.documents_action_download)) },
                onClick = {
                    isExpanded = false
                    onDownload()
                },
            )
        }
    }
}
