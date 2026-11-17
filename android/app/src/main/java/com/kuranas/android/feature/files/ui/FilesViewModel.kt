package com.kuranas.android.feature.files.ui

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.toDownloadRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

const val FOLDER_ID_ARGUMENT = "folderId"
const val FOLDER_NAME_ARGUMENT = "name"

data class FilesUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val files: List<FileItemDto> = emptyList(),
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val loadMoreError: String? = null,
    val breadcrumb: List<Pair<String, String>> = emptyList(),
    val error: String? = null,
    val showCreateFolderDialog: Boolean = false,
    val actionErrorMessage: String? = null,
    val fileAwaitingDeleteConfirmation: FileItemDto? = null,
    val selectedFileIds: Set<String> = emptySet(),
    val queuedDownloadsCount: Int = 0,
) {
    val isSelecting: Boolean get() = selectedFileIds.isNotEmpty()
}

@HiltViewModel
class FilesViewModel @Inject constructor(
    private val repository: FilesRepository,
    private val downloadQueue: DownloadQueue,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val filesLoader = PagedListLoader<FileItemDto>(viewModelScope) { it.id }
    private val navigationState = MutableStateFlow(FilesUiState())

    val state: StateFlow<FilesUiState> = combine(navigationState, filesLoader.state) { navigation, files ->
        navigation.copy(
            isLoading = files.isLoading,
            isRefreshing = files.isRefreshing,
            files = files.items,
            isLoadingMore = files.isLoadingMore,
            hasMore = files.hasMore,
            loadMoreError = files.loadMoreError,
            error = files.error,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, FilesUiState())

    init {
        val initialFolderId = savedStateHandle.get<String>(FOLDER_ID_ARGUMENT)
        if (initialFolderId == null) loadRoot() else openFolder(initialFolder(initialFolderId, savedStateHandle))
    }

    private fun initialFolder(folderId: String, savedStateHandle: SavedStateHandle) = FileItemDto(
        rawId = folderId.toIntOrNull() ?: 0,
        name = savedStateHandle.get<String>(FOLDER_NAME_ARGUMENT).orEmpty(),
        type = FileItemDto.TYPE_DIRECTORY,
    )

    fun loadRoot() {
        navigationState.update { it.copy(breadcrumb = emptyList()) }
        loadCurrentFolder()
    }

    fun openFolder(folder: FileItemDto) {
        navigationState.update { it.copy(breadcrumb = it.breadcrumb + (folder.name to folder.id)) }
        loadCurrentFolder()
    }

    fun navigateUp() {
        val breadcrumb = navigationState.value.breadcrumb
        if (breadcrumb.isEmpty()) return
        navigationState.update { it.copy(breadcrumb = breadcrumb.dropLast(1)) }
        loadCurrentFolder()
    }

    fun retry() = loadCurrentFolder()

    fun loadMore() = filesLoader.loadMore()

    fun retryLoadMore() = filesLoader.retryLoadMore()

    /** Reloads the current folder from the first page keeping the visible list (pull-to-refresh and resume). */
    fun refresh() = filesLoader.refresh()

    fun toggleSelection(file: FileItemDto) {
        if (file.isDir) return
        navigationState.update {
            val toggled = if (file.id in it.selectedFileIds) it.selectedFileIds - file.id else it.selectedFileIds + file.id
            it.copy(selectedFileIds = toggled)
        }
    }

    fun clearSelection() {
        navigationState.update { it.copy(selectedFileIds = emptySet()) }
    }

    fun downloadSelected() {
        val selectedIds = navigationState.value.selectedFileIds
        val selectedFiles = state.value.files.filter { it.id in selectedIds && !it.isDir }
        if (selectedFiles.isEmpty()) return
        downloadQueue.enqueue(selectedFiles.map { it.toDownloadRequest() })
        navigationState.update { it.copy(selectedFileIds = emptySet(), queuedDownloadsCount = selectedFiles.size) }
    }

    fun dismissQueuedDownloadsNotice() {
        navigationState.update { it.copy(queuedDownloadsCount = 0) }
    }

    private fun loadCurrentFolder() {
        navigationState.update { it.copy(selectedFileIds = emptySet()) }
        val currentFolderId = navigationState.value.breadcrumb.lastOrNull()?.second
        filesLoader.load { page ->
            if (currentFolderId == null) repository.getRootFiles(page) else repository.getChildrenById(currentFolderId, page)
        }
    }

    fun requestDelete(file: FileItemDto) {
        navigationState.update { it.copy(fileAwaitingDeleteConfirmation = file) }
    }

    fun cancelDelete() {
        navigationState.update { it.copy(fileAwaitingDeleteConfirmation = null) }
    }

    fun confirmDelete() {
        val confirmedFile = navigationState.value.fileAwaitingDeleteConfirmation ?: return
        navigationState.update { it.copy(fileAwaitingDeleteConfirmation = null) }
        viewModelScope.launch {
            when (val deletion = repository.deleteFile(confirmedFile.id)) {
                is AppResult.Success -> {
                    filesLoader.removeItem(confirmedFile.id)
                    filesLoader.refresh()
                }
                is AppResult.Error -> navigationState.update { it.copy(actionErrorMessage = deletion.message) }
            }
        }
    }

    fun toggleFileStarred(file: FileItemDto) {
        val currentIsStarred = state.value.files.firstOrNull { it.id == file.id }?.starred ?: return
        applyStarred(file.id, !currentIsStarred)
        viewModelScope.launch {
            when (val starring = repository.setFileStarred(file.id, !currentIsStarred)) {
                is AppResult.Success -> applyStarred(file.id, starring.data.starred)
                is AppResult.Error -> {
                    applyStarred(file.id, currentIsStarred)
                    navigationState.update { it.copy(actionErrorMessage = starring.message) }
                }
            }
        }
    }

    fun dismissActionError() {
        navigationState.update { it.copy(actionErrorMessage = null) }
    }

    private fun applyStarred(fileId: String, isStarred: Boolean) {
        val currentFile = state.value.files.firstOrNull { it.id == fileId } ?: return
        val index = filesLoader.removeItem(fileId)
        filesLoader.reinsertItem(index, currentFile.copy(starred = isStarred))
    }

    fun createFolder(name: String) {
        val parentId = navigationState.value.breadcrumb.lastOrNull()?.second
        viewModelScope.launch {
            repository.createFolder(name, parentId)
            navigationState.update { it.copy(showCreateFolderDialog = false) }
            filesLoader.refresh()
        }
    }

    fun toggleCreateFolderDialog(show: Boolean) {
        navigationState.update { it.copy(showCreateFolderDialog = show) }
    }
}
