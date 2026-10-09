package com.kuranas.android.feature.music.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.music.data.FIRST_PAGE
import com.kuranas.android.feature.music.data.MusicPage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class PagedListState<T>(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val isLoadingMore: Boolean = false,
    val items: List<T> = emptyList(),
    val nextPage: Int = FIRST_PAGE,
    val hasMore: Boolean = false,
    val error: String? = null,
)

class PagedListLoader<T>(
    private val scope: CoroutineScope,
    private val identityOf: (T) -> Any,
) {
    private val _state = MutableStateFlow(PagedListState<T>())
    val state: StateFlow<PagedListState<T>> = _state.asStateFlow()

    private var fetchPage: (suspend (Int) -> AppResult<MusicPage<T>>)? = null
    private var pageLoadJob: Job? = null

    fun load(fetchPage: suspend (Int) -> AppResult<MusicPage<T>>) {
        this.fetchPage = fetchPage
        restartFromFirstPage(
            onStart = { it.copy(isLoading = true, isRefreshing = false, isLoadingMore = false, error = null) },
            onError = { current, message -> current.copy(isLoading = false, error = message) },
        )
    }

    fun refresh() {
        if (fetchPage == null) return
        restartFromFirstPage(
            onStart = { it.copy(isRefreshing = true, isLoadingMore = false, error = null) },
            onError = { current, message -> current.copy(isRefreshing = false, error = message) },
        )
    }

    fun loadMore() {
        val fetch = fetchPage ?: return
        val current = _state.value
        val isBusy = current.isLoading || current.isRefreshing || current.isLoadingMore
        if (!current.hasMore || isBusy) return

        _state.update { it.copy(isLoadingMore = true) }
        pageLoadJob = scope.launch {
            when (val result = fetch(current.nextPage)) {
                is AppResult.Success -> _state.update { it.withAppendedPage(result.data) }
                is AppResult.Error -> _state.update { it.copy(isLoadingMore = false) }
            }
        }
    }

    private fun restartFromFirstPage(
        onStart: (PagedListState<T>) -> PagedListState<T>,
        onError: (PagedListState<T>, String) -> PagedListState<T>,
    ) {
        val fetch = fetchPage ?: return
        pageLoadJob?.cancel()
        _state.update(onStart)
        pageLoadJob = scope.launch {
            when (val result = fetch(FIRST_PAGE)) {
                is AppResult.Success -> _state.update { it.withFirstPage(result.data) }
                is AppResult.Error -> _state.update { onError(it, result.message) }
            }
        }
    }

    private fun PagedListState<T>.withFirstPage(page: MusicPage<T>) = copy(
        isLoading = false,
        isRefreshing = false,
        items = page.items.distinctBy(identityOf),
        nextPage = FIRST_PAGE + 1,
        hasMore = page.hasNext,
    )

    private fun PagedListState<T>.withAppendedPage(page: MusicPage<T>) = copy(
        isLoadingMore = false,
        items = (items + page.items).distinctBy(identityOf),
        nextPage = nextPage + 1,
        hasMore = page.hasNext,
    )
}
