package com.kuranas.android.feature.trash.ui

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
import androidx.compose.material.icons.filled.DeleteForever
import androidx.compose.material.icons.filled.DeleteSweep
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material.icons.filled.RestoreFromTrash
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
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
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.formatBytes
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.trash.data.TrashItemDto
import java.time.Instant

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TrashScreen(viewModel: TrashViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.items.size, state.hasMore, viewModel::loadMore)

    LaunchedEffect(Unit) { viewModel.loadTrash() }

    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(state.actionErrorMessage) {
        val message = state.actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissActionError()
    }

    val canEmptyTrash = state.items.isNotEmpty() && !state.isEmptyingTrash

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(
                title = stringResource(R.string.nav_trash),
                trailingIcon = if (canEmptyTrash) Icons.Default.DeleteSweep else null,
                trailingContentDescription = stringResource(R.string.trash_empty_action),
                onTrailingClick = viewModel::requestEmptyTrash,
            )
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                val errorMessage = state.error
                when {
                    state.isLoading || state.isEmptyingTrash -> LoadingView()
                    errorMessage != null -> ErrorView(message = errorMessage, onRetry = viewModel::loadTrash)
                    state.items.isEmpty() -> EmptyView(message = stringResource(R.string.trash_empty_state))
                    else -> LazyColumn(
                        state = listState,
                        contentPadding = PaddingValues(top = 8.dp, bottom = 80.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        items(state.items, key = { it.id }) { item ->
                            TrashListItem(
                                item = item,
                                onRestore = { viewModel.restore(item) },
                                onDeleteForever = { viewModel.requestDeleteForever(item) },
                            )
                        }
                    }
                }
            }
        }
    }

    state.pendingConfirmation?.let { confirmation ->
        TrashConfirmationDialog(
            confirmation = confirmation,
            onConfirm = viewModel::confirmPendingAction,
            onDismiss = viewModel::cancelConfirmation,
        )
    }
}

@Composable
private fun TrashConfirmationDialog(confirmation: TrashConfirmation, onConfirm: () -> Unit, onDismiss: () -> Unit) {
    val title = when (confirmation) {
        is TrashConfirmation.DeleteForever -> stringResource(R.string.trash_delete_forever_title, confirmation.item.name)
        TrashConfirmation.EmptyTrash -> stringResource(R.string.trash_empty_action)
    }
    val warning = when (confirmation) {
        is TrashConfirmation.DeleteForever -> stringResource(R.string.trash_delete_forever_warning)
        TrashConfirmation.EmptyTrash -> stringResource(R.string.trash_empty_warning)
    }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title) },
        text = { Text(warning) },
        confirmButton = {
            TextButton(onClick = onConfirm) {
                Text(stringResource(R.string.trash_delete_forever_action), color = MaterialTheme.colorScheme.error)
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text(stringResource(R.string.action_cancel)) }
        },
    )
}

@Composable
private fun TrashListItem(item: TrashItemDto, onRestore: () -> Unit, onDeleteForever: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(
            imageVector = if (item.isDirectory) Icons.Default.Folder else Icons.Default.InsertDriveFile,
            contentDescription = null,
            modifier = Modifier.size(24.dp),
            tint = if (item.isDirectory) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.secondary,
        )
        Column(modifier = Modifier.weight(1f)) {
            Text(item.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                stringResource(R.string.trash_origin, originFolderOf(item.originalPath)),
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(trashItemDetails(item), style = MaterialTheme.typography.bodySmall)
        }
        IconButton(onClick = onRestore, modifier = Modifier.size(36.dp)) {
            Icon(
                imageVector = Icons.Default.RestoreFromTrash,
                contentDescription = stringResource(R.string.trash_restore_action),
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(20.dp),
            )
        }
        IconButton(onClick = onDeleteForever, modifier = Modifier.size(36.dp)) {
            Icon(
                imageVector = Icons.Default.DeleteForever,
                contentDescription = stringResource(R.string.trash_delete_forever_action),
                tint = MaterialTheme.colorScheme.error,
                modifier = Modifier.size(20.dp),
            )
        }
    }
}

@Composable
private fun trashItemDetails(item: TrashItemDto): String {
    val occupiedSize = formatBytes(item.size)
    val remainingDays = daysUntilExpiry(item.expiresAt, Instant.now()) ?: return occupiedSize
    val expiry = if (remainingDays == 0) {
        stringResource(R.string.trash_expires_next_cleanup)
    } else {
        pluralStringResource(R.plurals.trash_expires_in_days, remainingDays, remainingDays)
    }
    return stringResource(R.string.trash_size_and_expiry, occupiedSize, expiry)
}
