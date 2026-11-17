package com.kuranas.android.core.discovery

import com.kuranas.android.core.server.ServerCheckResult
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.awaitCancellation
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.test.currentTime
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ServerSearchTest {

    private val serverA = DiscoveredServer("a", "192.168.0.10", 8000)
    private val serverB = DiscoveredServer("b", "192.168.0.11", 8000)

    private fun finiteSource(vararg servers: DiscoveredServer): () -> Flow<DiscoveredServer> =
        { flow { servers.forEach { emit(it) } } }

    private fun searchOf(
        sources: List<() -> Flow<DiscoveredServer>>,
        verdictsByUrl: Map<String, ServerCheckResult> = emptyMap(),
    ) = ServerSearch(
        candidateSources = sources,
        confirmCandidate = { url -> verdictsByUrl[url] ?: ServerCheckResult.Confirmed },
        searchTimeoutMillis = 10_000L,
    )

    @Test
    fun `emits confirmed candidate`() = runTest {
        val confirmedServers = searchOf(listOf(finiteSource(serverA))).search().toList()

        assertEquals(listOf(serverA), confirmedServers)
    }

    @Test
    fun `drops candidates that are not confirmed`() = runTest {
        val verdictsByUrl = mapOf(
            serverA.url to ServerCheckResult.NotKuraNas,
            serverB.url to ServerCheckResult.NoResponse,
        )

        val confirmedServers = searchOf(listOf(finiteSource(serverA, serverB)), verdictsByUrl).search().toList()

        assertEquals(emptyList<DiscoveredServer>(), confirmedServers)
    }

    @Test
    fun `emits repeated address once across sources`() = runTest {
        val confirmedServers = searchOf(listOf(finiteSource(serverA), finiteSource(serverA))).search().toList()

        assertEquals(listOf(serverA), confirmedServers)
    }

    @Test
    fun `finishes at ten seconds even when mdns never ends`() = runTest {
        val neverEndingSource: () -> Flow<DiscoveredServer> = {
            flow {
                emit(serverA)
                awaitCancellation()
            }
        }

        val confirmedServers = searchOf(listOf(neverEndingSource)).search().toList()

        assertEquals(listOf(serverA), confirmedServers)
        assertEquals(10_000L, currentTime)
    }

    @Test
    fun `keeps udp results when mdns fails`() = runTest {
        val failingSource: () -> Flow<DiscoveredServer> = { flow { throw IllegalStateException("nsd down") } }

        val confirmedServers = searchOf(listOf(failingSource, finiteSource(serverB))).search().toList()

        assertEquals(listOf(serverB), confirmedServers)
    }
}
