package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.PagedListFooter
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.UploadDestination

/** Tela do alvo de compartilhamento: escolha de destino e confirmação da fila de envio. */
@Composable
fun ShareTargetScreen(
    selectedFileCount: Int,
    onClose: () -> Unit,
    onViewTransfers: () -> Unit,
    uploadViewModel: UploadViewModel = hiltViewModel(),
) {
    val uploadState by uploadViewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp).navigationBarsPadding()) {
        KNHeader(title = stringResource(R.string.upload_share_title))
        when (val phase = uploadState.phase) {
            UploadPhase.Idle, UploadPhase.ResolvingShare -> LoadingView()
            is UploadPhase.ChoosingDestination -> DestinationChooser(
                selectedFileCount = selectedFileCount,
                lastDestination = phase.lastDestination,
                onConfirm = uploadViewModel::confirmDestination,
            )
            UploadPhase.Preparing -> UploadPreparingContent(modifier = Modifier.fillMaxWidth().padding(16.dp))
            is UploadPhase.Queued -> ShareResult(phase = phase, onViewTransfers = onViewTransfers, onClose = onClose)
        }
    }
}

@Composable
private fun ShareResult(phase: UploadPhase.Queued, onViewTransfers: () -> Unit, onClose: () -> Unit) {
    Column(modifier = Modifier.fillMaxWidth().padding(16.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
        UploadQueuedContent(phase)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            if (phase.queuedCount > 0) {
                Button(onClick = onViewTransfers) { Text(stringResource(R.string.upload_view_transfers)) }
            }
            TextButton(onClick = onClose) { Text(stringResource(R.string.upload_dismiss)) }
        }
    }
}

@Composable
private fun ColumnScope.DestinationChooser(
    selectedFileCount: Int,
    lastDestination: UploadDestination?,
    onConfirm: (UploadDestination, Boolean) -> Unit,
    viewModel: FilesViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var hasOpenedLastDestination by rememberSaveable { mutableStateOf(false) }
    var shouldAlwaysSendHere by rememberSaveable { mutableStateOf(false) }

    LaunchedEffect(lastDestination) {
        if (hasOpenedLastDestination) return@LaunchedEffect
        hasOpenedLastDestination = true
        if (lastDestination?.folderId == null) return@LaunchedEffect
        viewModel.openFolder(lastDestination.toFolderItem())
    }

    val currentFolder = state.breadcrumb.lastOrNull()
    val currentDestination = UploadDestination(currentFolder?.second, currentFolder?.first.orEmpty())
    val listState = remember { LazyListState() }
    val subfolders = state.files.filter { it.isDir }
    LoadMoreOnNearEnd(listState, subfolders.size, state.hasMore, viewModel::loadMore)

    Text(
        pluralStringResource(R.plurals.upload_share_selected_count, selectedFileCount, selectedFileCount),
        style = MaterialTheme.typography.bodyMedium,
        modifier = Modifier.padding(bottom = 8.dp),
    )
    if (currentFolder != null) {
        TextButton(onClick = viewModel::navigateUp) {
            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null, modifier = Modifier.size(18.dp))
            Text(stringResource(R.string.action_back), modifier = Modifier.padding(start = 8.dp))
        }
    }
    Box(modifier = Modifier.weight(1f).fillMaxWidth()) {
        when {
            state.isLoading -> LoadingView()
            state.error != null -> ErrorView(state.error!!, onRetry = viewModel::retry)
            else -> LazyColumn(state = listState, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                if (subfolders.isEmpty() && !state.hasMore) {
                    item { Text(stringResource(R.string.upload_share_no_subfolders), style = MaterialTheme.typography.bodySmall) }
                }
                items(subfolders, key = { it.id }) { folder -> FolderChoiceRow(folder, onClick = { viewModel.openFolder(folder) }) }
                item(key = "paged-list-footer") {
                    PagedListFooter(
                        isLoadingMore = state.isLoadingMore,
                        loadMoreError = state.loadMoreError,
                        onRetry = viewModel::retryLoadMore,
                    )
                }
            }
        }
    }
    Row(
        modifier = Modifier.fillMaxWidth().clickable { shouldAlwaysSendHere = !shouldAlwaysSendHere },
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Checkbox(checked = shouldAlwaysSendHere, onCheckedChange = { shouldAlwaysSendHere = it })
        Text(stringResource(R.string.upload_always_send_here))
    }
    Button(
        onClick = { onConfirm(currentDestination, shouldAlwaysSendHere) },
        modifier = Modifier.fillMaxWidth().padding(bottom = 16.dp),
    ) {
        Text(stringResource(R.string.upload_share_send_to, destinationLabel(currentDestination)), maxLines = 1, overflow = TextOverflow.Ellipsis)
    }
}

@Composable
private fun FolderChoiceRow(folder: FileItemDto, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(Icons.Default.Folder, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
        Text(folder.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
    }
}

private fun UploadDestination.toFolderItem() = FileItemDto(
    rawId = folderId?.toIntOrNull() ?: 0,
    name = folderName,
    type = FileItemDto.TYPE_DIRECTORY,
)
