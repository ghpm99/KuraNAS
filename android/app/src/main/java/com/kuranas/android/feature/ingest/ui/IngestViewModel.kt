package com.kuranas.android.feature.ingest.ui

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.ingest.data.IngestDownloadDto
import com.kuranas.android.feature.ingest.data.IngestPresetDto
import com.kuranas.android.feature.ingest.data.IngestRepository
import com.kuranas.android.feature.ingest.data.IngestTargetDto
import com.kuranas.android.navigation.AppRoute
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class IngestFormState(
    val url: String = "",
    val subfolder: String = "",
    val targets: List<IngestTargetDto> = emptyList(),
    val presets: List<IngestPresetDto> = emptyList(),
    val selectedTargetPath: String? = null,
    val selectedPresetKey: String? = null,
    val optionsError: String? = null,
    val isLoadingOptions: Boolean = true,
    val isSubmitting: Boolean = false,
    val submitError: String? = null,
) {
    val canSubmit: Boolean
        get() = url.isNotBlank() && selectedTargetPath != null && selectedPresetKey != null && !isSubmitting
}

data class IngestUiState(
    val form: IngestFormState = IngestFormState(),
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val downloads: List<IngestDownloadDto> = emptyList(),
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val loadMoreError: String? = null,
    val error: String? = null,
    val actionErrorMessage: String? = null,
)

@HiltViewModel
class IngestViewModel @Inject constructor(
    private val repository: IngestRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val downloadsLoader = PagedListLoader<IngestDownloadDto>(viewModelScope) { it.id }
    private val form = MutableStateFlow(IngestFormState(url = savedStateHandle.get<String>(AppRoute.INGEST_URL_ARGUMENT).orEmpty()))
    private val actionErrorMessage = MutableStateFlow<String?>(null)

    val state: StateFlow<IngestUiState> = combine(downloadsLoader.state, form, actionErrorMessage) { downloads, currentForm, actionError ->
        IngestUiState(
            form = currentForm,
            isLoading = downloads.isLoading,
            isRefreshing = downloads.isRefreshing,
            downloads = downloads.items,
            isLoadingMore = downloads.isLoadingMore,
            hasMore = downloads.hasMore,
            loadMoreError = downloads.loadMoreError,
            error = downloads.error,
            actionErrorMessage = actionError,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, IngestUiState())

    fun load() {
        loadOptions()
        loadDownloads()
    }

    fun loadDownloads() = downloadsLoader.load(repository::getDownloads)

    fun refresh() = downloadsLoader.refresh()

    fun loadMore() = downloadsLoader.loadMore()

    fun retryLoadMore() = downloadsLoader.retryLoadMore()

    fun loadOptions() {
        form.update { it.copy(isLoadingOptions = true, optionsError = null) }
        viewModelScope.launch {
            val targetsAnswer = repository.getTargets()
            val presetsAnswer = repository.getPresets()
            val optionsError = (targetsAnswer as? AppResult.Error)?.message ?: (presetsAnswer as? AppResult.Error)?.message
            val targets = (targetsAnswer as? AppResult.Success)?.data.orEmpty()
            val presets = (presetsAnswer as? AppResult.Success)?.data.orEmpty()
            form.update {
                it.copy(
                    targets = targets,
                    presets = presets,
                    selectedTargetPath = it.selectedTargetPath.takeIf { path -> targets.any { target -> target.path == path } }
                        ?: targets.firstOrNull()?.path,
                    selectedPresetKey = it.selectedPresetKey.takeIf { key -> presets.any { preset -> preset.key == key } }
                        ?: presets.firstOrNull()?.key,
                    optionsError = optionsError,
                    isLoadingOptions = false,
                )
            }
        }
    }

    fun changeUrl(url: String) = form.update { it.copy(url = url, submitError = null) }

    fun changeSubfolder(subfolder: String) = form.update { it.copy(subfolder = subfolder, submitError = null) }

    fun selectTarget(targetPath: String) = form.update { it.copy(selectedTargetPath = targetPath) }

    fun selectPreset(presetKey: String) = form.update { it.copy(selectedPresetKey = presetKey) }

    fun submit() {
        val current = form.value
        val targetPath = current.selectedTargetPath ?: return
        val presetKey = current.selectedPresetKey ?: return
        if (!current.canSubmit) return
        form.update { it.copy(isSubmitting = true, submitError = null) }
        viewModelScope.launch {
            val answer = repository.requestDownload(current.url.trim(), presetKey, targetPath, current.subfolder.trim())
            when (answer) {
                is AppResult.Success -> {
                    form.update { it.copy(isSubmitting = false, url = "", subfolder = "") }
                    refresh()
                }
                is AppResult.Error -> form.update { it.copy(isSubmitting = false, submitError = answer.message) }
            }
        }
    }

    fun cancel(download: IngestDownloadDto) {
        val jobId = download.jobId ?: return
        runDownloadAction { repository.cancelDownload(jobId) }
    }

    fun retry(download: IngestDownloadDto) = runDownloadAction { repository.retryDownload(download.id) }

    fun dismissActionError() {
        actionErrorMessage.value = null
    }

    private fun runDownloadAction(sendToServer: suspend () -> AppResult<Unit>) {
        viewModelScope.launch {
            val answer = sendToServer()
            actionErrorMessage.value = (answer as? AppResult.Error)?.message
            refresh()
        }
    }
}
