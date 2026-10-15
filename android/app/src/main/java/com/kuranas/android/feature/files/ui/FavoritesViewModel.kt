package com.kuranas.android.feature.files.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredKind
import com.kuranas.android.feature.files.data.StarredSort
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class FavoritesUiState(
    val kind: StarredKind = StarredKind.ALL,
    val sort: StarredSort = StarredSort.NAME,
    val isLoading: Boolean = true,
    val favorites: List<FileItemDto> = emptyList(),
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = false,
    val error: String? = null,
    val actionErrorMessage: String? = null,
)

@HiltViewModel
class FavoritesViewModel @Inject constructor(private val repository: FilesRepository) : ViewModel() {

    private val selection = MutableStateFlow(FavoritesSelection())
    private val actionErrorMessage = MutableStateFlow<String?>(null)
    private val favoritesLoader = PagedListLoader<FileItemDto>(viewModelScope) { it.id }

    val state: StateFlow<FavoritesUiState> = combine(
        selection,
        favoritesLoader.state,
        actionErrorMessage,
    ) { currentSelection, favorites, actionError ->
        FavoritesUiState(
            kind = currentSelection.kind,
            sort = currentSelection.sort,
            isLoading = favorites.isLoading,
            favorites = favorites.items,
            isLoadingMore = favorites.isLoadingMore,
            hasMore = favorites.hasMore,
            error = favorites.error,
            actionErrorMessage = actionError,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, FavoritesUiState())

    init {
        loadFirstPage()
    }

    fun selectKind(kind: StarredKind) {
        selection.update { it.copy(kind = kind) }
        loadFirstPage()
    }

    fun selectSort(sort: StarredSort) {
        selection.update { it.copy(sort = sort) }
        loadFirstPage()
    }

    fun loadMore() = favoritesLoader.loadMore()

    fun retry() = loadFirstPage()

    fun dismissActionError() {
        actionErrorMessage.value = null
    }

    fun unstar(file: FileItemDto) {
        val removedIndex = favoritesLoader.removeItem(file.id)
        if (removedIndex < 0) return
        viewModelScope.launch {
            val result = repository.setFileStarred(file.id, false)
            if (result !is AppResult.Error) return@launch
            favoritesLoader.reinsertItem(removedIndex, file)
            actionErrorMessage.value = result.message
        }
    }

    private fun loadFirstPage() {
        val (kind, sort) = selection.value
        favoritesLoader.load { page -> repository.getStarredFiles(page, kind, sort) }
    }
}

private data class FavoritesSelection(
    val kind: StarredKind = StarredKind.ALL,
    val sort: StarredSort = StarredSort.NAME,
)
