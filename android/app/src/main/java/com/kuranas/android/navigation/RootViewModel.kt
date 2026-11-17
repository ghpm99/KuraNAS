package com.kuranas.android.navigation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.server.ReconnectionOutcome
import com.kuranas.android.core.server.ServerReconnection
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RootViewModel @Inject constructor(
    private val serverStore: ServerStore,
    private val serverReconnection: ServerReconnection,
) : ViewModel() {

    private sealed interface Verification {
        data object Pending : Verification
        data class Finished(val unusableServer: ServerState?) : Verification
    }

    private val verification = MutableStateFlow<Verification>(Verification.Pending)

    val destination: StateFlow<RootDestination> =
        combine(serverStore.serverState, verification, ::destinationFor)
            .stateIn(viewModelScope, SharingStarted.Eagerly, RootDestination.Checking)

    init {
        viewModelScope.launch {
            val outcome = serverReconnection.ensureReachable()
            verification.value = Verification.Finished(unusableServerFor(outcome))
        }
    }

    private suspend fun unusableServerFor(outcome: ReconnectionOutcome): ServerState? = when (outcome) {
        ReconnectionOutcome.Unreachable, ReconnectionOutcome.AccessDenied -> serverStore.serverState.first()
        else -> null
    }

    private fun destinationFor(currentServer: ServerState, verification: Verification): RootDestination {
        if (verification !is Verification.Finished) return RootDestination.Checking
        if (currentServer == verification.unusableServer) return RootDestination.Connection
        return if (currentServer.isConnected) RootDestination.App else RootDestination.Connection
    }
}
