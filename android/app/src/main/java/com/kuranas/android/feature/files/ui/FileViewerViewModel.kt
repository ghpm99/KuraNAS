package com.kuranas.android.feature.files.ui

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileContent
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.canPreviewInline
import com.kuranas.android.feature.files.data.download.DownloadExecutor
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadState
import com.kuranas.android.feature.files.data.download.toDownloadRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface FileViewerState {
    data object Loading : FileViewerState
    data class Content(val content: FileContent, val info: FileItemDto) : FileViewerState
    data class TooLargeToPreview(val info: FileItemDto) : FileViewerState
    data class Failed(val message: String) : FileViewerState
}

sealed interface ExternalDelivery {
    data class Ready(val uri: Uri, val mimeType: String, val fileName: String, val purpose: DownloadPurpose) : ExternalDelivery
    data class Failed(val failure: DownloadFailure) : ExternalDelivery
}

@HiltViewModel
class FileViewerViewModel @Inject constructor(
    private val repository: FilesRepository,
    private val downloadQueue: DownloadQueue,
    private val downloadExecutor: DownloadExecutor,
) : ViewModel() {

    private val mutableState = MutableStateFlow<FileViewerState>(FileViewerState.Loading)
    val state: StateFlow<FileViewerState> = mutableState.asStateFlow()

    private val loadedFile = MutableStateFlow<FileItemDto?>(null)

    val downloadState: StateFlow<DownloadState?> = combine(downloadQueue.entries, loadedFile) { entries, file ->
        entries.lastOrNull { it.request.fileId == file?.id && it.request.purpose == DownloadPurpose.SAVE_TO_DEVICE }?.state
    }.stateIn(viewModelScope, SharingStarted.Eagerly, null)

    private val mutablePreparation = MutableStateFlow<DownloadProgress?>(null)
    val externalPreparation: StateFlow<DownloadProgress?> = mutablePreparation.asStateFlow()

    private val mutableDeliveries = MutableSharedFlow<ExternalDelivery>()
    val deliveries: SharedFlow<ExternalDelivery> = mutableDeliveries.asSharedFlow()

    fun load(fileId: String) {
        viewModelScope.launch {
            mutableState.value = FileViewerState.Loading
            when (val infoResult = repository.getFileInfo(fileId)) {
                is AppResult.Error -> mutableState.value = FileViewerState.Failed(infoResult.message)
                is AppResult.Success -> loadContent(infoResult.data)
            }
        }
    }

    private suspend fun loadContent(info: FileItemDto) {
        loadedFile.value = info
        if (!canPreviewInline(info.size)) {
            mutableState.value = FileViewerState.TooLargeToPreview(info)
            return
        }
        mutableState.value = when (val contentResult = repository.getFileContent(info.id)) {
            is AppResult.Success -> FileViewerState.Content(contentResult.data, info)
            is AppResult.Error -> FileViewerState.Failed(contentResult.message)
        }
    }

    fun download() {
        val file = loadedFile.value ?: return
        if (downloadState.value.isInProgress()) return
        viewModelScope.launch { downloadQueue.enqueue(listOf(file.toDownloadRequest(DownloadPurpose.SAVE_TO_DEVICE))) }
    }

    fun openWith() = deliverExternally(DownloadPurpose.OPEN_WITH)

    fun share() = deliverExternally(DownloadPurpose.SHARE)

    private fun deliverExternally(purpose: DownloadPurpose) {
        val file = loadedFile.value ?: return
        if (mutablePreparation.value != null) return
        val request = file.toDownloadRequest(purpose)
        viewModelScope.launch {
            mutablePreparation.value = DownloadProgress(receivedBytes = 0, totalBytes = request.sizeBytes)
            val outcome = try {
                downloadExecutor.execute(request) { mutablePreparation.value = it }
            } finally {
                mutablePreparation.value = null
            }
            mutableDeliveries.emit(
                when (outcome) {
                    is DownloadOutcome.Completed ->
                        ExternalDelivery.Ready(outcome.uri, outcome.mimeType, request.fileName, purpose)
                    is DownloadOutcome.Failed -> ExternalDelivery.Failed(outcome.failure)
                },
            )
        }
    }

    private fun DownloadState?.isInProgress() = this is DownloadState.Queued || this is DownloadState.Running
}
