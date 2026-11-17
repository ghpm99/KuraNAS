package com.kuranas.android.core.server

import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.discovery.ServerSearch
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.flow.emptyFlow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test

class ServerReconnectionTest {

    private val rememberedUrl = "http://192.168.0.10:8000"
    private val serverStore = mockk<ServerStore>(relaxed = true)
    private val healthChecker = mockk<ServerHealthChecker>()
    private val serverSearch = mockk<ServerSearch>()
    private val reconnection = ServerReconnection(serverStore, healthChecker, serverSearch)

    private fun rememberServer() {
        every { serverStore.serverUrl } returns flowOf(rememberedUrl)
    }

    @Test
    fun `sem servidor lembrado devolve NoServer sem consultar a rede`() = runTest {
        every { serverStore.serverUrl } returns flowOf(null)

        assertEquals(ReconnectionOutcome.NoServer, reconnection.ensureReachable())
        coVerify(exactly = 0) { healthChecker.check(any()) }
    }

    @Test
    fun `servidor lembrado que responde devolve Reachable`() = runTest {
        rememberServer()
        coEvery { healthChecker.check(rememberedUrl) } returns ServerCheckResult.Confirmed

        assertEquals(ReconnectionOutcome.Reachable, reconnection.ensureReachable())
        verify(exactly = 0) { serverSearch.search() }
    }

    @Test
    fun `acesso negado devolve AccessDenied sem buscar outro servidor`() = runTest {
        rememberServer()
        coEvery { healthChecker.check(rememberedUrl) } returns ServerCheckResult.AccessDenied

        assertEquals(ReconnectionOutcome.AccessDenied, reconnection.ensureReachable())
        verify(exactly = 0) { serverSearch.search() }
    }

    @Test
    fun `servidor sem resposta e busca que acha outro salva o novo e devolve Relocated`() = runTest {
        rememberServer()
        coEvery { healthChecker.check(rememberedUrl) } returns ServerCheckResult.NoResponse
        every { serverSearch.search() } returns flowOf(DiscoveredServer("nas", "192.168.0.20", 8000))

        assertEquals(ReconnectionOutcome.Relocated, reconnection.ensureReachable())
        coVerify { serverStore.saveDiscovered("http://192.168.0.20:8000", "nas") }
    }

    @Test
    fun `endereco que nao e KuraNAS e busca sem resultado devolve Unreachable`() = runTest {
        rememberServer()
        coEvery { healthChecker.check(rememberedUrl) } returns ServerCheckResult.NotKuraNas
        every { serverSearch.search() } returns emptyFlow()

        assertEquals(ReconnectionOutcome.Unreachable, reconnection.ensureReachable())
        coVerify(exactly = 0) { serverStore.saveDiscovered(any(), any()) }
    }
}
