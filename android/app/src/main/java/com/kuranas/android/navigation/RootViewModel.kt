package com.kuranas.android.navigation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.server.AccessDenial
import com.kuranas.android.core.server.AccessDeniedEvents
import com.kuranas.android.core.server.ChangeServerAction
import com.kuranas.android.core.server.ReconnectionOutcome
import com.kuranas.android.core.server.ServerReconnection
import com.kuranas.android.core.server.ServerSelectionEvents
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RootViewModel @Inject constructor(
    private val serverStore: ServerStore,
    private val serverReconnection: ServerReconnection,
    private val serverSelectionEvents: ServerSelectionEvents,
    private val accessDeniedEvents: AccessDeniedEvents,
    private val changeServerAction: ChangeServerAction,
) : ViewModel() {

    private sealed interface Verification {
        data object Pending : Verification
        data class Finished(val failedAtSelectionCount: Int?) : Verification
    }

    private val verification = MutableStateFlow<Verification>(Verification.Pending)
    private val _isRetrying = MutableStateFlow(false)

    val isRetrying: StateFlow<Boolean> = _isRetrying.asStateFlow()

    val destination: StateFlow<RootDestination> =
        combine(
            serverStore.serverState,
            verification,
            serverSelectionEvents.selectionCount,
            accessDeniedEvents.denial,
            ::destinationFor,
        ).stateIn(viewModelScope, SharingStarted.Eagerly, RootDestination.Checking)

    init {
        viewModelScope.launch {
            val outcome = serverReconnection.ensureReachable()
            applyOutcome(outcome)
            verification.value = Verification.Finished(failedAtSelectionCountFor(outcome))
        }
    }

    fun retryAfterAccessDenied() {
        if (_isRetrying.value) return
        _isRetrying.value = true
        viewModelScope.launch {
            try {
                val outcome = serverReconnection.ensureReachable()
                applyOutcome(outcome)
                verification.update { Verification.Finished(failedAtSelectionCountFor(outcome)) }
            } finally {
                _isRetrying.value = false
            }
        }
    }

    fun changeServerAfterAccessDenied() {
        viewModelScope.launch {
            changeServerAction.changeServer()
            accessDeniedEvents.clear()
        }
    }

    private fun applyOutcome(outcome: ReconnectionOutcome) {
        if (outcome is ReconnectionOutcome.AccessDenied) accessDeniedEvents.report(outcome.denial) else accessDeniedEvents.clear()
    }

    private fun failedAtSelectionCountFor(outcome: ReconnectionOutcome): Int? = when (outcome) {
        ReconnectionOutcome.Unreachable -> serverSelectionEvents.selectionCount.value
        else -> null
    }

    private fun destinationFor(
        currentServer: ServerState,
        verification: Verification,
        selectionCount: Int,
        denial: AccessDenial?,
    ): RootDestination {
        if (verification !is Verification.Finished) return RootDestination.Checking
        if (denial != null && currentServer.isConnected) return RootDestination.AccessDenied(denial)
        if (selectionCount == verification.failedAtSelectionCount) return RootDestination.Connection
        return if (currentServer.isConnected) RootDestination.App else RootDestination.Connection
    }
}
