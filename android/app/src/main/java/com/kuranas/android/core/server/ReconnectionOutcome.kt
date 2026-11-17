package com.kuranas.android.core.server

/** Desfecho da verificação do servidor lembrado ao abrir o app. */
sealed interface ReconnectionOutcome {
    data object NoServer : ReconnectionOutcome
    data object Reachable : ReconnectionOutcome
    data object Relocated : ReconnectionOutcome
    data object AccessDenied : ReconnectionOutcome
    data object Unreachable : ReconnectionOutcome
}
