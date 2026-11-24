package com.kuranas.android.core.server

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import kotlin.coroutines.cancellation.CancellationException

/**
 * Observa se o servidor lembrado responde, enquanto [start] estiver ativo. Verifica de imediato
 * quando a rede muda ou quando [checkNow] é chamado, e periodicamente conforme [policy].
 */
class ServerConnectionMonitor(
    private val serverUrl: Flow<String?>,
    private val networkChanges: Flow<Unit>,
    private val checkServer: suspend (String) -> ServerCheckResult,
    private val scope: CoroutineScope,
    private val policy: ConnectionCheckPolicy = ConnectionCheckPolicy(),
) {
    private val _state = MutableStateFlow(ServerConnectionState.Online)
    val state: StateFlow<ServerConnectionState> = _state.asStateFlow()

    private val _reconnections = MutableSharedFlow<Unit>(extraBufferCapacity = 1)
    val reconnections: SharedFlow<Unit> = _reconnections.asSharedFlow()

    private val checkRequests = Channel<Unit>(Channel.CONFLATED)
    private var snapshot = ConnectionSnapshot()
    private var monitoringJob: Job? = null

    fun start() {
        if (monitoringJob?.isActive == true) return
        monitoringJob = scope.launch {
            launch { collectNetworkChanges() }
            serverUrl.collectLatest { url -> monitorServer(url) }
        }
    }

    fun stop() {
        monitoringJob?.cancel()
        monitoringJob = null
    }

    fun checkNow() {
        checkRequests.trySend(Unit)
    }

    private suspend fun collectNetworkChanges() {
        try {
            networkChanges.collect { checkNow() }
        } catch (error: CancellationException) {
            throw error
        } catch (_: Exception) {
            return
        }
    }

    private suspend fun monitorServer(url: String?) {
        if (url == null) {
            apply(ConnectionObservation.NoServer)
            return
        }
        while (true) {
            apply(observe(url))
            val delayMillis = policy.nextCheckDelayMillis(snapshot)
            withTimeoutOrNull(delayMillis) { checkRequests.receive() }
        }
    }

    private suspend fun observe(url: String): ConnectionObservation = try {
        checkServer(url).toObservation()
    } catch (error: CancellationException) {
        throw error
    } catch (_: Exception) {
        ConnectionObservation.Unreachable
    }

    private fun apply(observation: ConnectionObservation) {
        val transition = reduceConnection(snapshot, observation)
        snapshot = transition.snapshot
        _state.value = transition.snapshot.state
        if (transition.hasReconnected) _reconnections.tryEmit(Unit)
    }
}
