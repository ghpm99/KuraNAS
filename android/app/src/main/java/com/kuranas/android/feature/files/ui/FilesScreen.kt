package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.AudioFile
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Image
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material.icons.filled.VideoFile
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
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
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.files.data.FileItemDto

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FilesScreen(
    onOpenImage: (String) -> Unit,
    onOpenVideo: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onOpenFile: (String, String) -> Unit,
    viewModel: FilesViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var newFolderName by remember { mutableStateOf("") }

    // Recarrega a pasta atual ao retomar a tela (voltar de outra tela / background),
    // pulando o primeiro ON_RESUME pra não duplicar o load do init.
    var firstResume by rememberSaveable { mutableStateOf(true) }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (firstResume) firstResume = false else viewModel.refresh()
    }

    Scaffold(
        containerColor = androidx.compose.ui.graphics.Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.toggleCreateFolderDialog(true) },
                containerColor = MaterialTheme.colorScheme.primary,
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.files_new_folder))
            }
        },
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(
                title = if (state.breadcrumb.isEmpty()) stringResource(R.string.nav_files) else state.breadcrumb.last().first,
                leadingIcon = if (state.breadcrumb.isNotEmpty()) Icons.AutoMirrored.Filled.ArrowBack else null,
                onLeadingClick = viewModel::navigateUp,
            )

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
                state.error != null -> ErrorView(state.error!!, onRetry = viewModel::loadRoot)
                state.files.isEmpty() -> EmptyView(stringResource(R.string.files_empty_folder))
                else -> LazyColumn(
                    contentPadding = PaddingValues(bottom = 80.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    items(state.files, key = { it.id }) { file ->
                        FileListItem(
                            file = file,
                            onClick = {
                                when {
                                    file.isDir -> viewModel.openFolder(file)
                                    file.mimeType.startsWith("image/") -> onOpenImage(file.id)
                                    file.mimeType.startsWith("video/") -> onOpenVideo(file.id)
                                    file.mimeType.startsWith("audio/") -> onPlayAudio(file.id)
                                    else -> onOpenFile(file.id, file.name)
                                }
                            },
                            onStar = { viewModel.starFile(file) },
                        )
                    }
                }
            }
            }
        }
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
private fun FileListItem(file: FileItemDto, onClick: () -> Unit, onStar: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
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
                contentDescription = null,
                tint = if (file.isStarred) MaterialTheme.colorScheme.secondary else MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.size(18.dp),
            )
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
