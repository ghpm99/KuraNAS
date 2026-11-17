package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.formatBytes
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.DownloadState
import com.kuranas.android.feature.files.data.download.toMessage

@Composable
fun DownloadStatusView(state: DownloadState?, modifier: Modifier = Modifier) {
    when (state) {
        null -> Unit
        DownloadState.Queued -> Text(
            stringResource(R.string.download_status_queued),
            style = MaterialTheme.typography.bodySmall,
            modifier = modifier,
        )
        is DownloadState.Running -> ProgressStatus(state.progress, modifier)
        is DownloadState.Completed -> Text(
            stringResource(R.string.download_status_completed),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.primary,
            modifier = modifier,
        )
        is DownloadState.Failed -> Text(
            state.failure.toMessage(LocalContext.current),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.error,
            modifier = modifier,
        )
    }
}

@Composable
fun ProgressStatus(progress: DownloadProgress, modifier: Modifier = Modifier) {
    Column(modifier = modifier) {
        val fraction = progress.fraction
        if (fraction == null) {
            LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
        } else {
            LinearProgressIndicator(progress = { fraction }, modifier = Modifier.fillMaxWidth())
        }
        Text(
            stringResource(
                R.string.download_status_progress,
                formatBytes(progress.receivedBytes),
                formatBytes(progress.totalBytes),
            ),
            style = MaterialTheme.typography.bodySmall,
        )
    }
}
