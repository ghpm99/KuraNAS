package com.kuranas.android.feature.files.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class FilesUiState(
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val files: List<FileItemDto> = emptyList(),
    val breadcrumb: List<Pair<String, String>> = emptyList(),
    val error: String? = null,
    val showCreateFolderDialog: Boolean = false,
)

@HiltViewModel
class FilesViewModel @Inject constructor(private val repository: FilesRepository) : ViewModel() {

    private val _state = MutableStateFlow(FilesUiState(isLoading = true))
    val state: StateFlow<FilesUiState> = _state.asStateFlow()

    init { loadRoot() }

    fun loadRoot() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null, breadcrumb = emptyList()) }
            when (val r = repository.getRootFiles()) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, files = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun openFolder(folder: FileItemDto) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            when (val r = repository.getChildrenById(folder.id)) {
                is AppResult.Success -> _state.update { state ->
                    state.copy(
                        isLoading = false,
                        files = r.data,
                        breadcrumb = state.breadcrumb + (folder.name to folder.id),
                    )
                }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun navigateUp() {
        val crumb = _state.value.breadcrumb
        if (crumb.isEmpty()) return
        val newCrumb = crumb.dropLast(1)
        if (newCrumb.isEmpty()) {
            loadRoot()
            return
        }
        val parentId = newCrumb.last().second
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, breadcrumb = newCrumb) }
            when (val r = repository.getChildrenById(parentId)) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, files = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun deleteFile(file: FileItemDto) {
        viewModelScope.launch {
            repository.deleteFile(file.id)
            refreshCurrent()
        }
    }

    fun starFile(file: FileItemDto) {
        viewModelScope.launch { repository.starFile(file.id) }
    }

    fun createFolder(name: String) {
        val parentId = _state.value.breadcrumb.lastOrNull()?.second
        viewModelScope.launch {
            repository.createFolder(name, parentId)
            _state.update { it.copy(showCreateFolderDialog = false) }
            refreshCurrent()
        }
    }

    fun toggleCreateFolderDialog(show: Boolean) {
        _state.update { it.copy(showCreateFolderDialog = show) }
    }

    private fun refreshCurrent() {
        val crumb = _state.value.breadcrumb
        if (crumb.isEmpty()) loadRoot()
        else viewModelScope.launch {
            when (val r = repository.getChildrenById(crumb.last().second)) {
                is AppResult.Success -> _state.update { it.copy(files = r.data) }
                is AppResult.Error -> _state.update { it.copy(error = r.message) }
            }
        }
    }

    /**
     * Recarrega a pasta atual mantendo a lista visível (sem spinner de tela cheia).
     * Usado pelo pull-to-refresh e pelo refetch ao retomar a tela.
     */
    fun refresh() {
        val crumb = _state.value.breadcrumb
        viewModelScope.launch {
            _state.update { it.copy(isRefreshing = true, error = null) }
            val result = if (crumb.isEmpty()) repository.getRootFiles()
            else repository.getChildrenById(crumb.last().second)
            when (result) {
                is AppResult.Success -> _state.update { it.copy(isRefreshing = false, files = result.data) }
                is AppResult.Error -> _state.update { it.copy(isRefreshing = false, error = result.message) }
            }
        }
    }
}
