package com.kuranas.android.core.server

enum class ServerConnectionState { Online, Offline }

sealed interface ConnectionObservation {
    data object Reachable : ConnectionObservation
    data object Unreachable : ConnectionObservation
    data object NoServer : ConnectionObservation
}

data class ConnectionSnapshot(
    val state: ServerConnectionState = ServerConnectionState.Online,
    val consecutiveFailures: Int = 0,
)

data class ConnectionTransition(val snapshot: ConnectionSnapshot, val hasReconnected: Boolean)

fun ServerCheckResult.toObservation(): ConnectionObservation = when (this) {
    ServerCheckResult.Confirmed -> ConnectionObservation.Reachable
    ServerCheckResult.NoResponse, ServerCheckResult.NotKuraNas, ServerCheckResult.AccessDenied -> ConnectionObservation.Unreachable
}

fun reduceConnection(current: ConnectionSnapshot, observation: ConnectionObservation): ConnectionTransition = when (observation) {
    ConnectionObservation.NoServer -> ConnectionTransition(ConnectionSnapshot(), hasReconnected = false)
    ConnectionObservation.Unreachable -> ConnectionTransition(
        ConnectionSnapshot(ServerConnectionState.Offline, current.consecutiveFailures + 1),
        hasReconnected = false,
    )
    ConnectionObservation.Reachable -> ConnectionTransition(
        ConnectionSnapshot(ServerConnectionState.Online, consecutiveFailures = 0),
        hasReconnected = current.state == ServerConnectionState.Offline,
    )
}

data class ConnectionCheckPolicy(
    val onlineIntervalMillis: Long = 60_000L,
    val offlineInitialIntervalMillis: Long = 10_000L,
    val offlineStepMillis: Long = 10_000L,
    val offlineMaxIntervalMillis: Long = 30_000L,
) {
    fun nextCheckDelayMillis(snapshot: ConnectionSnapshot): Long {
        if (snapshot.state == ServerConnectionState.Online) return onlineIntervalMillis
        val backedOff = offlineInitialIntervalMillis + offlineStepMillis * (snapshot.consecutiveFailures - 1).coerceAtLeast(0)
        return backedOff.coerceAtMost(offlineMaxIntervalMillis)
    }
}
