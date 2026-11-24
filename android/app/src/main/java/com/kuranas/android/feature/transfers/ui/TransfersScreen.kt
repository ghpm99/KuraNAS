package com.kuranas.android.feature.transfers.ui

import androidx.annotation.StringRes
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ClearAll
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Upload
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferStatus
import com.kuranas.android.ui.theme.StatusNegative

@Composable
fun TransfersScreen(
    onNavigateBack: () -> Unit,
    viewModel: TransfersViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = stringResource(R.string.transfers_title),
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )
        if (state.isEmpty) {
            EmptyView(stringResource(R.string.transfers_empty), icon = Icons.Default.Upload)
            return@Column
        }
        TextButton(onClick = viewModel::clearFinished, enabled = state.hasFinishedItems) {
            Icon(Icons.Default.ClearAll, contentDescription = null, modifier = Modifier.size(18.dp))
            Text(stringResource(R.string.transfers_clear_finished), modifier = Modifier.padding(start = 8.dp))
        }
        LazyColumn(contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            section(R.string.transfers_section_in_progress, state.inProgress, viewModel)
            section(R.string.transfers_section_failed, state.failed, viewModel)
            section(R.string.transfers_section_completed, state.completed, viewModel)
        }
    }
}

private fun androidx.compose.foundation.lazy.LazyListScope.section(
    @StringRes titleRes: Int,
    transfers: List<TransferItem>,
    viewModel: TransfersViewModel,
) {
    if (transfers.isEmpty()) return
    item(key = "header-$titleRes") {
        Text(
            stringResource(titleRes),
            style = MaterialTheme.typography.labelLarge,
            modifier = Modifier.padding(top = 8.dp),
        )
    }
    items(transfers, key = { it.id }) { transfer ->
        TransferCard(
            transfer = transfer,
            onCancel = { viewModel.cancel(transfer.id) },
            onRetry = { viewModel.retry(transfer.id) },
        )
    }
}

@Composable
private fun TransferCard(transfer: TransferItem, onCancel: () -> Unit, onRetry: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 16.dp)
            .padding(16.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Icon(
                imageVector = if (transfer.direction == TransferDirection.UPLOAD) Icons.Default.Upload else Icons.Default.Download,
                contentDescription = null,
                modifier = Modifier.size(18.dp),
            )
            Text(transfer.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, modifier = Modifier.weight(1f))
        }
        transfer.destinationFolderName?.let { folderName ->
            Text(stringResource(R.string.transfers_destination, folderName), style = MaterialTheme.typography.bodySmall)
        }
        Text(
            text = transferStatusText(transfer),
            style = MaterialTheme.typography.bodySmall,
            color = if (transfer.status == TransferStatus.FAILED) StatusNegative else MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (transfer.status == TransferStatus.RUNNING) {
            Spacer(Modifier.height(8.dp))
            val progressFraction = transfer.progressFraction
            if (progressFraction == null) {
                LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
            } else {
                LinearProgressIndicator(progress = { progressFraction }, modifier = Modifier.fillMaxWidth())
            }
        }
        if (transfer.canCancel) {
            TextButton(onClick = onCancel) { Text(stringResource(R.string.action_cancel)) }
        }
        if (transfer.canRetry) {
            TextButton(onClick = onRetry) { Text(stringResource(R.string.transfers_action_retry)) }
        }
    }
}

@Composable
private fun transferStatusText(transfer: TransferItem): String {
    val percent = ((transfer.progressFraction ?: 0f) * 100).toInt()
    return when (transfer.status) {
        TransferStatus.QUEUED -> stringResource(R.string.transfers_status_queued)
        TransferStatus.RUNNING -> when {
            transfer.progressFraction == null && transfer.direction == TransferDirection.UPLOAD -> stringResource(R.string.transfers_status_uploading_unknown)
            transfer.progressFraction == null -> stringResource(R.string.transfers_status_downloading_unknown)
            transfer.direction == TransferDirection.UPLOAD -> stringResource(R.string.transfers_status_uploading, percent)
            else -> stringResource(R.string.transfers_status_downloading, percent)
        }
        TransferStatus.WAITING_FOR_NETWORK -> stringResource(R.string.transfers_status_waiting_for_network)
        TransferStatus.COMPLETED -> stringResource(R.string.transfers_status_completed)
        TransferStatus.ALREADY_EXISTED -> stringResource(R.string.transfers_status_already_existed)
        TransferStatus.FAILED -> transfer.failureMessage?.let { stringResource(R.string.transfers_status_failed_with_reason, it) }
            ?: stringResource(R.string.transfers_status_failed)
        TransferStatus.CANCELLED -> stringResource(R.string.transfers_status_cancelled)
    }
}
