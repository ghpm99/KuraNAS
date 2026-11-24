package com.kuranas.android.feature.transfers.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TransfersViewModel @Inject constructor(
    transferDao: TransferDao,
    private val transferRepository: TransferRepository,
) : ViewModel() {

    val state: StateFlow<TransfersUiState> = transferDao.observeAll()
        .map(::groupTransfers)
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(STOP_TIMEOUT_MILLIS), TransfersUiState())

    fun cancel(transferId: Long) {
        viewModelScope.launch { transferRepository.cancel(transferId) }
    }

    fun retry(transferId: Long) {
        viewModelScope.launch { transferRepository.retry(transferId) }
    }

    fun clearFinished() {
        viewModelScope.launch { transferRepository.clearFinished() }
    }

    private companion object {
        const val STOP_TIMEOUT_MILLIS = 5_000L
    }
}
