package com.kuranas.android.feature.files.ui

import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.WindowInsetsSides
import androidx.compose.foundation.layout.only
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.AudioFile
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.RadioButtonUnchecked
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Image
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material.icons.filled.Upload
import androidx.compose.material.icons.filled.VideoFile
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SmallFloatingActionButton
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.activity.compose.LocalActivity
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.FileSizeText
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.PagedListFooter
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.UploadDestination

private const val ANY_MIME_TYPE = "*/*"

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FilesScreen(
    onOpenImage: (String) -> Unit,
    onOpenVideo: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onOpenFile: (String, String) -> Unit,
    onOpenTransfers: () -> Unit,
    viewModel: FilesViewModel = hiltViewModel(),
    uploadViewModel: UploadViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val uploadState by uploadViewModel.state.collectAsStateWithLifecycle()
    val filePicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { pickedUris ->
        if (pickedUris.isEmpty()) return@rememberLauncherForActivityResult
        val openFolder = state.breadcrumb.lastOrNull()
        uploadViewModel.sendPicked(pickedUris, UploadDestination(openFolder?.second, openFolder?.first.orEmpty()))
    }
    var newFolderName by remember { mutableStateOf("") }
    BackHandler(enabled = state.breadcrumb.isNotEmpty(), onBack = viewModel::navigateUp)
    val listState = rememberSaveable(state.breadcrumb.lastOrNull()?.second, saver = LazyListState.Saver) { LazyListState() }
    LoadMoreOnNearEnd(listState, state.files.size, state.hasMore, viewModel::loadMore)

    val activity = LocalActivity.current
    var shouldSkipNextResumeRefresh by rememberSaveable { mutableStateOf(true) }
    LifecycleEventEffect(Lifecycle.Event.ON_PAUSE) {
        if (activity?.isChangingConfigurations == true) shouldSkipNextResumeRefresh = true
    }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (shouldSkipNextResumeRefresh) shouldSkipNextResumeRefresh = false else viewModel.refresh()
    }

    val snackbarHostState = remember { SnackbarHostState() }
    val queuedDownloadsMessage = if (state.queuedDownloadsCount > 0) {
        pluralStringResource(R.plurals.files_downloads_queued, state.queuedDownloadsCount, state.queuedDownloadsCount)
    } else {
        null
    }
    LaunchedEffect(queuedDownloadsMessage) {
        val message = queuedDownloadsMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissQueuedDownloadsNotice()
    }
    LaunchedEffect(state.actionErrorMessage) {
        val message = state.actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissActionError()
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        containerColor = androidx.compose.ui.graphics.Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
        floatingActionButton = {
            Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                SmallFloatingActionButton(
                    onClick = { filePicker.launch(arrayOf(ANY_MIME_TYPE)) },
                    containerColor = MaterialTheme.colorScheme.secondary,
                ) {
                    Icon(Icons.Default.Upload, contentDescription = stringResource(R.string.upload_action))
                }
                FloatingActionButton(
                    onClick = { viewModel.toggleCreateFolderDialog(true) },
                    containerColor = MaterialTheme.colorScheme.primary,
                ) {
                    Icon(Icons.Default.Add, contentDescription = stringResource(R.string.files_new_folder))
                }
            }
        },
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(
                title = if (state.breadcrumb.isEmpty()) stringResource(R.string.nav_files) else state.breadcrumb.last().first,
                leadingIcon = if (state.breadcrumb.isNotEmpty()) Icons.AutoMirrored.Filled.ArrowBack else null,
                onLeadingClick = viewModel::navigateUp,
            )

            if (state.isSelecting) {
                SelectionBar(
                    selectedCount = state.selectedFileIds.size,
                    onDownload = viewModel::downloadSelected,
                    onClear = viewModel::clearSelection,
                )
            }

            if (state.breadcrumb.isNotEmpty()) {
                LazyRow(horizontalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.padding(bottom = 8.dp)) {
                    item { Text(stringResource(R.string.files_root), style = MaterialTheme.typography.bodySmall, modifier = Modifier.clickable { viewModel.loadRoot() }) }
                    items(state.breadcrumb) { (name, _) ->
                        Text(stringResource(R.string.files_breadcrumb_segment, name), style = MaterialTheme.typography.bodySmall)
                    }
                }
            }

            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
            when {
                state.isLoading -> LoadingView()
                state.error != null -> ErrorView(state.error!!, onRetry = viewModel::retry)
                state.files.isEmpty() -> EmptyView(stringResource(R.string.files_empty_folder))
                else -> LazyColumn(
                    state = listState,
                    contentPadding = PaddingValues(bottom = 80.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    items(state.files, key = { it.id }) { file ->
                        FileListItem(
                            file = file,
                            onClick = {
                                when {
                                    state.isSelecting -> viewModel.toggleSelection(file)
                                    file.isDir -> viewModel.openFolder(file)
                                    else -> openFileByType(file, onOpenImage, onOpenVideo, onPlayAudio, onOpenFile)
                                }
                            },
                            onStar = { viewModel.toggleFileStarred(file) },
                            onDelete = { viewModel.requestDelete(file) },
                            selection = FileSelection(
                                isSelecting = state.isSelecting,
                                isSelected = file.id in state.selectedFileIds,
                                onLongClick = { viewModel.toggleSelection(file) },
                            ),
                        )
                    }
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
        }
    }

    UploadDialog(
        phase = uploadState.phase,
        onViewTransfers = {
            uploadViewModel.dismissResult()
            onOpenTransfers()
        },
        onDismiss = uploadViewModel::dismissResult,
    )

    state.fileAwaitingDeleteConfirmation?.let { fileToDelete ->
        MoveToTrashDialog(
            fileName = fileToDelete.name,
            onConfirm = viewModel::confirmDelete,
            onDismiss = viewModel::cancelDelete,
        )
    }

    if (state.showCreateFolderDialog) {
        AlertDialog(
            onDismissRequest = { viewModel.toggleCreateFolderDialog(false) },
            title = { Text(stringResource(R.string.files_new_folder)) },
            text = {
                OutlinedTextField(
                    value = newFolderName,
                    onValueChange = { newFolderName = it },
                    label = { Text(stringResource(R.string.files_folder_name_label)) },
                    singleLine = true,
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    viewModel.createFolder(newFolderName)
                    newFolderName = ""
                }) { Text(stringResource(R.string.action_create)) }
            },
            dismissButton = {
                TextButton(onClick = { viewModel.toggleCreateFolderDialog(false) }) { Text(stringResource(R.string.action_cancel)) }
            },
        )
    }
}

@Composable
private fun SelectionBar(selectedCount: Int, onDownload: () -> Unit, onClear: () -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(
            stringResource(R.string.files_selection_count, selectedCount),
            style = MaterialTheme.typography.bodyMedium,
            modifier = Modifier.weight(1f),
        )
        TextButton(onClick = onDownload) {
            Icon(Icons.Default.Download, contentDescription = null, modifier = Modifier.size(18.dp))
            Spacer(Modifier.width(6.dp))
            Text(stringResource(R.string.files_selection_download))
        }
        IconButton(onClick = onClear) {
            Icon(Icons.Default.Close, contentDescription = stringResource(R.string.files_selection_clear))
        }
    }
}

@Composable
private fun MoveToTrashDialog(fileName: String, onConfirm: () -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.files_delete_title, fileName)) },
        text = { Text(stringResource(R.string.files_delete_goes_to_trash)) },
        confirmButton = {
            TextButton(onClick = onConfirm) { Text(stringResource(R.string.files_delete_confirm)) }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text(stringResource(R.string.action_cancel)) }
        },
    )
}

internal data class FileSelection(val isSelecting: Boolean, val isSelected: Boolean, val onLongClick: () -> Unit)

@OptIn(ExperimentalFoundationApi::class)
@Composable
internal fun FileListItem(
    file: FileItemDto,
    onClick: () -> Unit,
    onStar: () -> Unit,
    onDelete: (() -> Unit)? = null,
    selection: FileSelection? = null,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .combinedClickable(onClick = onClick, onLongClick = selection?.onLongClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (selection?.isSelecting == true && !file.isDir) {
            Icon(
                imageVector = if (selection.isSelected) Icons.Default.CheckCircle else Icons.Default.RadioButtonUnchecked,
                contentDescription = null,
                modifier = Modifier.size(24.dp),
                tint = MaterialTheme.colorScheme.primary,
            )
        }
        Icon(
            imageVector = fileIcon(file),
            contentDescription = null,
            modifier = Modifier.size(24.dp),
            tint = if (file.isDir) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.secondary,
        )
        Column(modifier = Modifier.weight(1f)) {
            Text(file.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
            if (!file.isDir) FileSizeText(file.size)
        }
        IconButton(onClick = onStar, modifier = Modifier.size(32.dp)) {
            Icon(
                imageVector = if (file.isStarred) Icons.Default.Star else Icons.Default.StarBorder,
                contentDescription = stringResource(if (file.isStarred) R.string.files_unstar_action else R.string.files_star_action),
                tint = if (file.isStarred) MaterialTheme.colorScheme.secondary else MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.size(18.dp),
            )
        }
        if (onDelete != null) {
            IconButton(onClick = onDelete, modifier = Modifier.size(32.dp)) {
                Icon(
                    imageVector = Icons.Default.Delete,
                    contentDescription = stringResource(R.string.files_delete_action),
                    tint = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.size(18.dp),
                )
            }
        }
    }
}

private fun fileIcon(file: FileItemDto): ImageVector = when {
    file.isDir -> Icons.Default.Folder
    file.mimeType.startsWith("image/") -> Icons.Default.Image
    file.mimeType.startsWith("video/") -> Icons.Default.VideoFile
    file.mimeType.startsWith("audio/") -> Icons.Default.AudioFile
    else -> Icons.Default.InsertDriveFile
}

internal fun openFileByType(
    file: FileItemDto,
    onOpenImage: (String) -> Unit,
    onOpenVideo: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onOpenFile: (String, String) -> Unit,
) {
    when {
        file.mimeType.startsWith("image/") -> onOpenImage(file.id)
        file.mimeType.startsWith("video/") -> onOpenVideo(file.id)
        file.mimeType.startsWith("audio/") -> onPlayAudio(file.id)
        else -> onOpenFile(file.id, file.name)
    }
}
