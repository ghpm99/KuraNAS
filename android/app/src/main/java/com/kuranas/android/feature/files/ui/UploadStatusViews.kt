package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.ui.unit.dp
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import com.kuranas.android.R
import com.kuranas.android.feature.files.data.UploadDestination

@Composable
internal fun destinationLabel(destination: UploadDestination): String =
    if (destination.isRoot) stringResource(R.string.files_root) else destination.folderName

@Composable
private fun uploadFailureText(reason: UploadFailureReason): String = when (reason) {
    UploadFailureReason.NoServer -> stringResource(R.string.upload_failure_no_server)
    is UploadFailureReason.ServerMessage -> reason.text
}

@Composable
internal fun UploadProgressContent(progress: UploadPhase.Sending, modifier: Modifier = Modifier) {
    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(stringResource(R.string.upload_progress_count, progress.current, progress.total))
        Text(progress.fileName, style = MaterialTheme.typography.bodySmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
        LinearProgressIndicator(
            progress = { (progress.current - 1).toFloat() / progress.total },
            modifier = Modifier.fillMaxWidth(),
        )
    }
}

@Composable
internal fun UploadSummaryContent(summary: UploadSummary, modifier: Modifier = Modifier) {
    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(4.dp)) {
        if (summary.uploadedCount > 0) Text(stringResource(R.string.upload_summary_uploaded, summary.uploadedCount))
        if (summary.renamedCount > 0) Text(stringResource(R.string.upload_summary_renamed, summary.renamedCount))
        if (summary.skippedCount > 0) Text(stringResource(R.string.upload_summary_skipped, summary.skippedCount))
        if (summary.hasFailures) {
            Text(stringResource(R.string.upload_summary_failed, summary.failures.size), color = MaterialTheme.colorScheme.error)
            summary.failures.forEach { failure ->
                Text(
                    stringResource(R.string.upload_failure_line, failure.file.name, uploadFailureText(failure.reason)),
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

/** Diálogo de envio do explorer: progresso enquanto envia e resumo ao final, com "tentar de novo" para as falhas. */
@Composable
internal fun UploadDialog(phase: UploadPhase, onRetryFailed: () -> Unit, onDismiss: () -> Unit) {
    when (phase) {
        is UploadPhase.Sending -> AlertDialog(
            onDismissRequest = {},
            title = { Text(stringResource(R.string.upload_action)) },
            text = { UploadProgressContent(phase) },
            confirmButton = {},
        )
        is UploadPhase.Finished -> AlertDialog(
            onDismissRequest = onDismiss,
            title = { Text(stringResource(R.string.upload_summary_title)) },
            text = { UploadSummaryContent(phase.summary) },
            confirmButton = {
                if (phase.summary.hasFailures) {
                    TextButton(onClick = onRetryFailed) { Text(stringResource(R.string.upload_retry_failed)) }
                }
            },
            dismissButton = { TextButton(onClick = onDismiss) { Text(stringResource(R.string.action_close)) } },
        )
        else -> Unit
    }
}
