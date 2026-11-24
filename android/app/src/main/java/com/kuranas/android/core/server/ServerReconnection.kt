package com.kuranas.android.core.server

import com.kuranas.android.core.discovery.ServerSearch
import kotlinx.coroutines.flow.firstOrNull
import javax.inject.Inject

/** Garante que o servidor lembrado responde, procurando-o na rede quando mudou de endereço. */
class ServerReconnection @Inject constructor(
    private val serverStore: ServerStore,
    private val healthChecker: ServerHealthChecker,
    private val serverSearch: ServerSearch,
) {

    /**
     * Valida o servidor lembrado. Se não responde ou deixou de ser KuraNAS, busca na rede e
     * salva o primeiro servidor confirmado. Acesso negado não dispara busca.
     */
    suspend fun ensureReachable(): ReconnectionOutcome {
        val rememberedUrl = serverStore.serverUrl.firstOrNull() ?: return ReconnectionOutcome.NoServer
        return when (val result = healthChecker.check(rememberedUrl)) {
            ServerCheckResult.Confirmed -> ReconnectionOutcome.Reachable
            is ServerCheckResult.AccessDenied -> ReconnectionOutcome.AccessDenied(result.denial)
            ServerCheckResult.NoResponse, ServerCheckResult.NotKuraNas -> relocate()
        }
    }

    private suspend fun relocate(): ReconnectionOutcome {
        val foundServer = serverSearch.search().firstOrNull() ?: return ReconnectionOutcome.Unreachable
        serverStore.saveDiscovered(foundServer.url, foundServer.name)
        return ReconnectionOutcome.Relocated
    }
}
