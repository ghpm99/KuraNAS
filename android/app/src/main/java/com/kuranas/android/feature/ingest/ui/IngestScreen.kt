package com.kuranas.android.feature.ingest.ui

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
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Button
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.paging.LoadMoreOnNearEnd
import com.kuranas.android.core.ui.ReconnectionRefresh
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.PagedListFooter
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.ingest.data.IngestDownloadDto
import com.kuranas.android.feature.ingest.data.canBeCanceled
import com.kuranas.android.feature.ingest.data.canBeRetried

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun IngestScreen(onNavigateBack: () -> Unit, viewModel: IngestViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, state.downloads.size, state.hasMore, viewModel::loadMore)

    LaunchedEffect(Unit) { viewModel.load() }
    ReconnectionRefresh(viewModel::load)

    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(state.actionErrorMessage) {
        val message = state.actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissActionError()
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
    ) { innerPadding ->
        Column(modifier = Modifier.fillMaxSize().padding(innerPadding).padding(horizontal = 16.dp)) {
            KNHeader(
                title = stringResource(R.string.ingest_title),
                leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
                onLeadingClick = onNavigateBack,
            )
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                LazyColumn(
                    state = listState,
                    contentPadding = PaddingValues(top = 8.dp, bottom = 80.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    item {
                        IngestForm(
                            form = state.form,
                            onUrlChange = viewModel::changeUrl,
                            onSubfolderChange = viewModel::changeSubfolder,
                            onTargetSelected = viewModel::selectTarget,
                            onPresetSelected = viewModel::selectPreset,
                            onRetryOptions = viewModel::loadOptions,
                            onSubmit = viewModel::submit,
                        )
                    }
                    item {
                        Text(
                            text = stringResource(R.string.ingest_downloads_heading),
                            style = MaterialTheme.typography.titleMedium,
                            modifier = Modifier.padding(top = 12.dp),
                        )
                    }
                    downloadsSection(state, viewModel)
                }
            }
        }
    }
}

private fun androidx.compose.foundation.lazy.LazyListScope.downloadsSection(state: IngestUiState, viewModel: IngestViewModel) {
    val errorMessage = state.error
    when {
        state.isLoading -> item { LoadingView(modifier = Modifier.padding(24.dp)) }
        errorMessage != null -> item { ErrorWithRetry(errorMessage, viewModel::loadDownloads) }
        state.downloads.isEmpty() -> item {
            Text(stringResource(R.string.ingest_downloads_empty), style = MaterialTheme.typography.bodyMedium)
        }
        else -> {
            items(state.downloads, key = { it.id }) { download ->
                IngestDownloadRow(
                    download = download,
                    onCancel = { viewModel.cancel(download) },
                    onRetry = { viewModel.retry(download) },
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

@Composable
private fun ErrorWithRetry(message: String, onRetry: () -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(message, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
        TextButton(onClick = onRetry) { Text(stringResource(R.string.action_retry)) }
    }
}

@Composable
private fun IngestForm(
    form: IngestFormState,
    onUrlChange: (String) -> Unit,
    onSubfolderChange: (String) -> Unit,
    onTargetSelected: (String) -> Unit,
    onPresetSelected: (String) -> Unit,
    onRetryOptions: () -> Unit,
    onSubmit: () -> Unit,
) {
    Column(
        modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light, radius = 12.dp).padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        OutlinedTextField(
            value = form.url,
            onValueChange = onUrlChange,
            label = { Text(stringResource(R.string.ingest_url_label)) },
            singleLine = true,
            isError = form.submitError != null,
            supportingText = form.submitError?.let { message -> { Text(message) } },
            modifier = Modifier.fillMaxWidth(),
        )
        val optionsError = form.optionsError
        when {
            form.isLoadingOptions -> LoadingView(modifier = Modifier.padding(8.dp))
            optionsError != null -> ErrorWithRetry(optionsError, onRetryOptions)
            form.targets.isEmpty() -> Text(stringResource(R.string.ingest_no_targets), style = MaterialTheme.typography.bodyMedium)
            else -> {
                OptionDropdown(
                    label = stringResource(R.string.ingest_target_label),
                    options = form.targets.map { it.path to it.label },
                    selectedKey = form.selectedTargetPath,
                    onSelected = onTargetSelected,
                )
                OutlinedTextField(
                    value = form.subfolder,
                    onValueChange = onSubfolderChange,
                    label = { Text(stringResource(R.string.ingest_subfolder_label)) },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )
                OptionDropdown(
                    label = stringResource(R.string.ingest_preset_label),
                    options = form.presets.map { it.key to it.label },
                    selectedKey = form.selectedPresetKey,
                    onSelected = onPresetSelected,
                )
            }
        }
        Button(onClick = onSubmit, enabled = form.canSubmit, modifier = Modifier.fillMaxWidth()) {
            Text(stringResource(R.string.ingest_submit_action))
        }
    }
}

@Composable
private fun OptionDropdown(
    label: String,
    options: List<Pair<String, String>>,
    selectedKey: String?,
    onSelected: (String) -> Unit,
) {
    var isExpanded by remember { mutableStateOf(false) }
    val selectedLabel = options.firstOrNull { it.first == selectedKey }?.second.orEmpty()
    Column(modifier = Modifier.fillMaxWidth()) {
        Text(label, style = MaterialTheme.typography.labelMedium)
        OutlinedButton(onClick = { isExpanded = true }, modifier = Modifier.fillMaxWidth()) {
            Text(selectedLabel, maxLines = 1, overflow = TextOverflow.Ellipsis)
        }
        DropdownMenu(expanded = isExpanded, onDismissRequest = { isExpanded = false }) {
            options.forEach { (optionKey, optionLabel) ->
                DropdownMenuItem(
                    text = { Text(optionLabel) },
                    onClick = {
                        isExpanded = false
                        onSelected(optionKey)
                    },
                )
            }
        }
    }
}

@Composable
private fun IngestDownloadRow(download: IngestDownloadDto, onCancel: () -> Unit, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light, radius = 12.dp).padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(
            text = download.title ?: download.url,
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            text = ingestDestinationLabel(download),
            style = MaterialTheme.typography.bodySmall,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically) {
            Text(
                text = ingestStateLabel(download),
                style = MaterialTheme.typography.bodySmall,
                modifier = Modifier.weight(1f),
            )
            if (download.canBeCanceled()) {
                TextButton(onClick = onCancel) { Text(stringResource(R.string.ingest_cancel_action)) }
            }
            if (download.canBeRetried()) {
                TextButton(onClick = onRetry) { Text(stringResource(R.string.ingest_retry_action)) }
            }
        }
        if (download.state == RUNNING_STATE) {
            LinearProgressIndicator(progress = { download.progress / PERCENT_SCALE }, modifier = Modifier.fillMaxWidth())
        }
        if (download.error.isNotBlank()) {
            Text(download.error, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.error)
        }
    }
}

@Composable
private fun ingestDestinationLabel(download: IngestDownloadDto): String {
    if (download.subfolder.isBlank()) return download.targetRootName
    return stringResource(R.string.ingest_destination, download.targetRootName, download.subfolder)
}

@Composable
private fun ingestStateLabel(download: IngestDownloadDto): String = when (download.state) {
    "queued" -> stringResource(R.string.ingest_state_queued)
    RUNNING_STATE -> stringResource(R.string.ingest_state_running, download.progress)
    "completed" -> stringResource(R.string.ingest_state_completed)
    "failed" -> stringResource(R.string.ingest_state_failed)
    "canceled" -> stringResource(R.string.ingest_state_canceled)
    else -> download.state
}

private const val RUNNING_STATE = "running"
private const val PERCENT_SCALE = 100f
