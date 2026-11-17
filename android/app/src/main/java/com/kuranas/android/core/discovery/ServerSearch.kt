package com.kuranas.android.core.discovery

import com.kuranas.android.core.server.ServerCheckResult
import com.kuranas.android.core.server.ServerHealthChecker
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.channelFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Busca servidores KuraNAS na rede combinando mDNS e broadcast UDP.
 * Só emite candidatos confirmados por `GET /api/v1/health`, uma vez por endereço.
 */
@Singleton
class ServerSearch internal constructor(
    private val candidateSources: List<() -> Flow<DiscoveredServer>>,
    private val confirmCandidate: suspend (String) -> ServerCheckResult,
    private val searchTimeoutMillis: Long,
) {

    @Inject
    constructor(
        nsdDiscovery: NsdDiscovery,
        udpDiscovery: UdpDiscovery,
        healthChecker: ServerHealthChecker,
    ) : this(
        candidateSources = listOf(nsdDiscovery::discover, udpDiscovery::discover),
        confirmCandidate = healthChecker::check,
        searchTimeoutMillis = SEARCH_TIMEOUT_MILLIS,
    )

    /**
     * Emite cada servidor confirmado. O fluxo termina quando todas as fontes terminam
     * ou após 10 segundos, o que ocorrer primeiro. Falha de uma fonte não afeta as outras.
     */
    fun search(): Flow<DiscoveredServer> = channelFlow {
        val seenUrls = ConcurrentHashMap.newKeySet<String>()
        withTimeoutOrNull(searchTimeoutMillis) {
            coroutineScope {
                candidateSources.forEach { openSource ->
                    launch {
                        openSource().catch { }.collect { candidate ->
                            if (!seenUrls.add(candidate.url)) return@collect
                            launch {
                                if (confirmCandidate(candidate.url) == ServerCheckResult.Confirmed) send(candidate)
                            }
                        }
                    }
                }
            }
        }
    }

    companion object {
        const val SEARCH_TIMEOUT_MILLIS = 10_000L
    }
}
