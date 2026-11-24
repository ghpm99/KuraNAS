package com.kuranas.android.feature.documents.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.documents.data.DocumentsRepository
import com.kuranas.android.feature.files.ui.ExternalDelivery
import com.kuranas.android.feature.files.data.download.DownloadExecutor
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

const val SEARCH_DEBOUNCE_MILLIS = 300L
const val SEARCH_MIN_LENGTH = 2

data class DocumentsUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val documents: List<DocumentListItem> = emptyList(),
    val totalDocuments: Int? = null,
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val loadMoreError: String? = null,
    val error: String? = null,
    val formatFilter: DocumentFormatFilter = DocumentFormatFilter.ALL,
    val searchText: String = "",
    val isSearching: Boolean = false,
    val preparingDocumentId: Int? = null,
)

sealed interface DocumentsEvent {
    data class OpenInViewer(val documentId: Int, val documentName: String) : DocumentsEvent
    data class OpenExternally(val delivery: ExternalDelivery) : DocumentsEvent
    data object DownloadQueued : DocumentsEvent
}

@HiltViewModel
class DocumentsViewModel @Inject constructor(
    private val repository: DocumentsRepository,
    private val downloadQueue: DownloadQueue,
    private val downloadExecutor: DownloadExecutor,
) : ViewModel() {

    private val documentsLoader = PagedListLoader<DocumentListItem>(viewModelScope) { it.id }
    private val totalDocuments = MutableStateFlow<Int?>(null)
    private val formatFilter = MutableStateFlow(DocumentFormatFilter.ALL)
    private val searchText = MutableStateFlow("")
    private val activeQuery = MutableStateFlow("")
    private val preparingDocumentId = MutableStateFlow<Int?>(null)
    private var searchDebounceJob: Job? = null
    private var hasStarted = false

    private val mutableEvents = MutableSharedFlow<DocumentsEvent>(extraBufferCapacity = 1)
    val events: SharedFlow<DocumentsEvent> = mutableEvents.asSharedFlow()

    private val controls = combine(formatFilter, searchText, activeQuery, preparingDocumentId) { filter, text, query, preparingId ->
        DocumentsControls(filter, text, query.isNotEmpty(), preparingId)
    }

    val state: StateFlow<DocumentsUiState> = combine(documentsLoader.state, totalDocuments, controls) { documents, total, controls ->
        DocumentsUiState(
            isLoading = documents.isLoading,
            isRefreshing = documents.isRefreshing,
            documents = documents.items,
            totalDocuments = total.takeUnless { controls.isSearching },
            isLoadingMore = documents.isLoadingMore,
            hasMore = documents.hasMore,
            loadMoreError = documents.loadMoreError,
            error = documents.error,
            formatFilter = controls.formatFilter,
            searchText = controls.searchText,
            isSearching = controls.isSearching,
            preparingDocumentId = controls.preparingDocumentId,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, DocumentsUiState())

    fun startIfNeeded() {
        if (hasStarted) return
        hasStarted = true
        reload()
    }

    fun reload() {
        val format = formatFilter.value.apiValue
        val query = activeQuery.value
        documentsLoader.load { page -> fetchPage(query, page, format) }
        loadTotal(format)
    }

    fun refresh() {
        documentsLoader.refresh()
        loadTotal(formatFilter.value.apiValue)
    }

    fun loadMore() = documentsLoader.loadMore()

    fun retryLoadMore() = documentsLoader.retryLoadMore()

    fun selectFormat(filter: DocumentFormatFilter) {
        if (filter == formatFilter.value) return
        formatFilter.value = filter
        reload()
    }

    fun changeSearchText(text: String) {
        searchText.value = text
        searchDebounceJob?.cancel()
        val query = text.trim()
        if (query.length < SEARCH_MIN_LENGTH) {
            applyQuery("")
            return
        }
        searchDebounceJob = viewModelScope.launch {
            delay(SEARCH_DEBOUNCE_MILLIS)
            applyQuery(query)
        }
    }

    fun clearSearch() = changeSearchText("")

    fun open(document: DocumentListItem) {
        when (documentOpenMode(document.format)) {
            DocumentOpenMode.IN_APP_VIEWER -> mutableEvents.tryEmit(DocumentsEvent.OpenInViewer(document.id, document.name))
            DocumentOpenMode.OPEN_WITH -> prepareForOtherApp(document)
        }
    }

    fun download(document: DocumentListItem) {
        viewModelScope.launch {
            downloadQueue.enqueue(listOf(document.toDownloadRequest(DownloadPurpose.SAVE_TO_DEVICE)))
            mutableEvents.emit(DocumentsEvent.DownloadQueued)
        }
    }

    private fun applyQuery(query: String) {
        if (query == activeQuery.value) return
        activeQuery.value = query
        reload()
    }

    private fun prepareForOtherApp(document: DocumentListItem) {
        if (preparingDocumentId.value != null) return
        val request = document.toDownloadRequest(DownloadPurpose.OPEN_WITH)
        preparingDocumentId.value = document.id
        viewModelScope.launch {
            val outcome = try {
                downloadExecutor.execute(request) {}
            } finally {
                preparingDocumentId.value = null
            }
            val delivery = when (outcome) {
                is DownloadOutcome.Completed -> ExternalDelivery.Ready(outcome.uri, outcome.mimeType, request.fileName, request.purpose)
                is DownloadOutcome.Failed -> ExternalDelivery.Failed(outcome.failure)
            }
            mutableEvents.emit(DocumentsEvent.OpenExternally(delivery))
        }
    }

    private suspend fun fetchPage(query: String, page: Int, format: String?): AppResult<ListPage<DocumentListItem>> {
        if (query.isEmpty()) {
            return when (val answer = repository.getDocuments(page, format)) {
                is AppResult.Success -> AppResult.Success(ListPage(answer.data.items.map { it.toListItem() }, answer.data.hasNext))
                is AppResult.Error -> answer
            }
        }
        return when (val answer = repository.searchDocuments(query, page, format)) {
            is AppResult.Success -> AppResult.Success(ListPage(answer.data.items.map { it.toListItem() }, answer.data.hasNext))
            is AppResult.Error -> answer
        }
    }

    private fun loadTotal(format: String?) {
        viewModelScope.launch {
            val answer = repository.getDocumentsCount(format)
            totalDocuments.value = (answer as? AppResult.Success)?.data
        }
    }

    private fun DocumentListItem.toDownloadRequest(purpose: DownloadPurpose) =
        DownloadRequest(fileId = id.toString(), fileName = name, sizeBytes = size, purpose = purpose)

    private data class DocumentsControls(
        val formatFilter: DocumentFormatFilter,
        val searchText: String,
        val isSearching: Boolean,
        val preparingDocumentId: Int?,
    )
}
