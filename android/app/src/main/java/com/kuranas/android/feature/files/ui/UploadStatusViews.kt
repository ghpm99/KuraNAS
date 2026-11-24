package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.kuranas.android.R
import com.kuranas.android.feature.files.data.UploadDestination

@Composable
internal fun destinationLabel(destination: UploadDestination): String =
    if (destination.isRoot) stringResource(R.string.files_root) else destination.folderName

@Composable
internal fun UploadPreparingContent(modifier: Modifier = Modifier) {
    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(stringResource(R.string.upload_preparing))
        LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
    }
}

@Composable
internal fun UploadQueuedContent(phase: UploadPhase.Queued, modifier: Modifier = Modifier) {
    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(4.dp)) {
        if (phase.queuedCount > 0) {
            Text(pluralStringResource(R.plurals.upload_queued_count, phase.queuedCount, phase.queuedCount))
        } else {
            Text(stringResource(R.string.upload_nothing_queued))
        }
        if (phase.unreadableFileNames.isNotEmpty()) {
            Text(stringResource(R.string.upload_unreadable_title), color = MaterialTheme.colorScheme.error)
            phase.unreadableFileNames.forEach { fileName ->
                Text(
                    stringResource(R.string.upload_unreadable_line, fileName.ifBlank { stringResource(R.string.upload_unnamed_file) }),
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

/** Diálogo do explorer: indica o preparo dos arquivos e confirma a entrada na fila, com atalho para Transferências. */
@Composable
internal fun UploadDialog(phase: UploadPhase, onViewTransfers: () -> Unit, onDismiss: () -> Unit) {
    when (phase) {
        UploadPhase.Preparing -> AlertDialog(
            onDismissRequest = {},
            title = { Text(stringResource(R.string.upload_action)) },
            text = { UploadPreparingContent() },
            confirmButton = {},
        )
        is UploadPhase.Queued -> AlertDialog(
            onDismissRequest = onDismiss,
            title = { Text(stringResource(R.string.upload_action)) },
            text = { UploadQueuedContent(phase) },
            confirmButton = {
                if (phase.queuedCount > 0) {
                    TextButton(onClick = onViewTransfers) { Text(stringResource(R.string.upload_view_transfers)) }
                }
            },
            dismissButton = { TextButton(onClick = onDismiss) { Text(stringResource(R.string.upload_dismiss)) } },
        )
        else -> Unit
    }
}
