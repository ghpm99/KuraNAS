package com.kuranas.android.feature.trash.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.PagedListLoader
import com.kuranas.android.feature.trash.data.TrashItemDto
import com.kuranas.android.feature.trash.data.TrashRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface TrashConfirmation {
    data class DeleteForever(val item: TrashItemDto) : TrashConfirmation
    data object EmptyTrash : TrashConfirmation
}

data class TrashUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val items: List<TrashItemDto> = emptyList(),
    val hasMore: Boolean = false,
    val error: String? = null,
    val actionErrorMessage: String? = null,
    val pendingConfirmation: TrashConfirmation? = null,
    val isEmptyingTrash: Boolean = false,
)

@HiltViewModel
class TrashViewModel @Inject constructor(private val repository: TrashRepository) : ViewModel() {

    private val trashLoader = PagedListLoader<TrashItemDto>(viewModelScope) { it.id }
    private val interaction = MutableStateFlow(TrashInteraction())

    val state: StateFlow<TrashUiState> = combine(trashLoader.state, interaction) { trash, currentInteraction ->
        TrashUiState(
            isLoading = trash.isLoading,
            isRefreshing = trash.isRefreshing,
            items = trash.items,
            hasMore = trash.hasMore,
            error = trash.error,
            actionErrorMessage = currentInteraction.actionErrorMessage,
            pendingConfirmation = currentInteraction.pendingConfirmation,
            isEmptyingTrash = currentInteraction.isEmptyingTrash,
        )
    }.stateIn(viewModelScope, SharingStarted.Eagerly, TrashUiState())

    fun loadTrash() = trashLoader.load(repository::getTrashItems)

    fun refresh() = trashLoader.refresh()

    fun loadMore() = trashLoader.loadMore()

    fun restore(item: TrashItemDto) = takeOutOfTrash(item) { repository.restoreItem(item.id) }

    fun requestDeleteForever(item: TrashItemDto) {
        interaction.update { it.copy(pendingConfirmation = TrashConfirmation.DeleteForever(item)) }
    }

    fun requestEmptyTrash() {
        interaction.update { it.copy(pendingConfirmation = TrashConfirmation.EmptyTrash) }
    }

    fun cancelConfirmation() {
        interaction.update { it.copy(pendingConfirmation = null) }
    }

    fun confirmPendingAction() {
        val confirmedAction = interaction.value.pendingConfirmation ?: return
        interaction.update { it.copy(pendingConfirmation = null) }
        when (confirmedAction) {
            is TrashConfirmation.DeleteForever ->
                takeOutOfTrash(confirmedAction.item) { repository.deleteItemForever(confirmedAction.item.id) }
            TrashConfirmation.EmptyTrash -> emptyWholeTrash()
        }
    }

    fun dismissActionError() {
        interaction.update { it.copy(actionErrorMessage = null) }
    }

    private fun takeOutOfTrash(item: TrashItemDto, sendToServer: suspend () -> AppResult<Unit>) {
        val removedIndex = trashLoader.removeItem(item.id)
        if (removedIndex < 0) return
        viewModelScope.launch {
            val serverAnswer = sendToServer()
            if (serverAnswer !is AppResult.Error) return@launch
            trashLoader.reinsertItem(removedIndex, item)
            interaction.update { it.copy(actionErrorMessage = serverAnswer.message) }
        }
    }

    private fun emptyWholeTrash() {
        interaction.update { it.copy(isEmptyingTrash = true) }
        viewModelScope.launch {
            val serverAnswer = repository.emptyTrash()
            interaction.update {
                it.copy(
                    isEmptyingTrash = false,
                    actionErrorMessage = (serverAnswer as? AppResult.Error)?.message,
                )
            }
            loadTrash()
        }
    }
}

private data class TrashInteraction(
    val actionErrorMessage: String? = null,
    val pendingConfirmation: TrashConfirmation? = null,
    val isEmptyingTrash: Boolean = false,
)
