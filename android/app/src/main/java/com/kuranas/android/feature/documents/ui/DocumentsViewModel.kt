package com.kuranas.android.feature.documents.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.documents.data.DocumentDto
import com.kuranas.android.feature.documents.data.DocumentsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

data class DocumentsUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val documents: List<DocumentDto> = emptyList(),
    val totalDocuments: Int? = null,
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val loadMoreError: String? = null,
    val error: String? = null,
)

@HiltViewModel
class DocumentsViewModel @Inject constructor(private val repository: DocumentsRepository) : ViewModel() {

    private val documentsLoader = PagedListLoader<DocumentDto>(viewModelScope) { it.id }
    private val totalDocuments = MutableStateFlow<Int?>(null)
    private var hasStarted = false

    val state: StateFlow<DocumentsUiState> = combine(documentsLoader.state, totalDocuments) { documents, total ->
        DocumentsUiState(
            isLoading = documents.isLoading,
            isRefreshing = documents.isRefreshing,
            documents = documents.items,
            totalDocuments = total,
            isLoadingMore = documents.isLoadingMore,
            hasMore = documents.hasMore,
            loadMoreError = documents.loadMoreError,
            error = documents.error,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, DocumentsUiState())

    fun startIfNeeded() {
        if (hasStarted) return
        hasStarted = true
        reload()
    }

    fun reload() {
        documentsLoader.load { page -> repository.getDocuments(page) }
        loadTotal()
    }

    fun refresh() {
        documentsLoader.refresh()
        loadTotal()
    }

    fun loadMore() = documentsLoader.loadMore()

    fun retryLoadMore() = documentsLoader.retryLoadMore()

    private fun loadTotal() {
        viewModelScope.launch {
            val answer = repository.getDocumentsCount()
            totalDocuments.value = (answer as? AppResult.Success)?.data
        }
    }
}
