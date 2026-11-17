package com.kuranas.android.navigation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.server.ReconnectionOutcome
import com.kuranas.android.core.server.ServerReconnection
import com.kuranas.android.core.server.ServerSelectionEvents
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RootViewModel @Inject constructor(
    private val serverStore: ServerStore,
    private val serverReconnection: ServerReconnection,
    private val serverSelectionEvents: ServerSelectionEvents,
) : ViewModel() {

    private sealed interface Verification {
        data object Pending : Verification
        data class Finished(val failedAtSelectionCount: Int?) : Verification
    }

    private val verification = MutableStateFlow<Verification>(Verification.Pending)

    val destination: StateFlow<RootDestination> =
        combine(serverStore.serverState, verification, serverSelectionEvents.selectionCount, ::destinationFor)
            .stateIn(viewModelScope, SharingStarted.Eagerly, RootDestination.Checking)

    init {
        viewModelScope.launch {
            val outcome = serverReconnection.ensureReachable()
            verification.value = Verification.Finished(failedAtSelectionCountFor(outcome))
        }
    }

    private fun failedAtSelectionCountFor(outcome: ReconnectionOutcome): Int? = when (outcome) {
        ReconnectionOutcome.Unreachable, ReconnectionOutcome.AccessDenied -> serverSelectionEvents.selectionCount.value
        else -> null
    }

    private fun destinationFor(
        currentServer: ServerState,
        verification: Verification,
        selectionCount: Int,
    ): RootDestination {
        if (verification !is Verification.Finished) return RootDestination.Checking
        if (selectionCount == verification.failedAtSelectionCount) return RootDestination.Connection
        return if (currentServer.isConnected) RootDestination.App else RootDestination.Connection
    }
}
