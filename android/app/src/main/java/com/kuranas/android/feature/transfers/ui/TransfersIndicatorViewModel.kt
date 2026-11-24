package com.kuranas.android.feature.transfers.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.transfers.data.TransferDao
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

@HiltViewModel
class TransfersIndicatorViewModel @Inject constructor(
    transferDao: TransferDao,
) : ViewModel() {

    val activeCount: StateFlow<Int> = transferDao.observeActiveCount()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(STOP_TIMEOUT_MILLIS), 0)

    private companion object {
        const val STOP_TIMEOUT_MILLIS = 5_000L
    }
}
