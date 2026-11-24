package com.kuranas.android.core.server

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.emptyFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ServerConnectionMonitorTest {

    private val serverUrl = MutableStateFlow<String?>("http://nas:8000")
    private val networkChanges = MutableSharedFlow<Unit>()
    private var nextResult: ServerCheckResult = ServerCheckResult.Confirmed
    private var checkCount = 0
    private val accessDeniedEvents = AccessDeniedEvents()

    private fun TestScope.monitor(changes: Flow<Unit> = networkChanges) = ServerConnectionMonitor(
        serverUrl = serverUrl,
        networkChanges = changes,
        checkServer = {
            checkCount++
            nextResult
        },
        scope = backgroundScope,
        accessDeniedEvents = accessDeniedEvents,
    )

    @Test
    fun `servidor que nao responde deixa o estado offline`() = runTest {
        nextResult = ServerCheckResult.NoResponse
        val monitor = monitor()

        monitor.start()
        runCurrent()

        assertEquals(ServerConnectionState.Offline, monitor.state.value)
    }

    @Test
    fun `acesso negado publica a recusa em vez de ficar offline`() = runTest {
        val denial = AccessDenial("10.0.0.9", "blocked")
        nextResult = ServerCheckResult.AccessDenied(denial)
        val monitor = monitor()

        monitor.start()
        runCurrent()

        assertEquals(ServerConnectionState.AccessDenied, monitor.state.value)
        assertEquals(denial, accessDeniedEvents.denial.value)
    }

    @Test
    fun `voltar a responder limpa a recusa`() = runTest {
        nextResult = ServerCheckResult.AccessDenied(AccessDenial("10.0.0.9", null))
        val monitor = monitor()
        monitor.start()
        runCurrent()

        nextResult = ServerCheckResult.Confirmed
        monitor.checkNow()
        runCurrent()

        assertEquals(ServerConnectionState.Online, monitor.state.value)
        assertEquals(null, accessDeniedEvents.denial.value)
    }

    @Test
    fun `sem servidor lembrado nao verifica e permanece online`() = runTest {
        serverUrl.value = null
        val monitor = monitor()

        monitor.start()
        runCurrent()

        assertEquals(ServerConnectionState.Online, monitor.state.value)
        assertEquals(0, checkCount)
    }

    @Test
    fun `volta a responder emite uma unica reconexao`() = runTest {
        nextResult = ServerCheckResult.NoResponse
        val monitor = monitor()
        var reconnectionCount = 0
        backgroundScope.launch { monitor.reconnections.collect { reconnectionCount++ } }
        monitor.start()
        runCurrent()

        nextResult = ServerCheckResult.Confirmed
        advanceTimeBy(10_001)
        runCurrent()
        advanceTimeBy(60_001)
        runCurrent()

        assertEquals(ServerConnectionState.Online, monitor.state.value)
        assertEquals(1, reconnectionCount)
    }

    @Test
    fun `perda de rede dispara verificacao imediata`() = runTest {
        val monitor = monitor()
        monitor.start()
        runCurrent()
        nextResult = ServerCheckResult.NoResponse

        networkChanges.emit(Unit)
        runCurrent()

        assertEquals(ServerConnectionState.Offline, monitor.state.value)
        assertEquals(2, checkCount)
    }

    @Test
    fun `tentar agora forca verificacao antes do intervalo`() = runTest {
        val monitor = monitor(emptyFlow())
        monitor.start()
        runCurrent()
        nextResult = ServerCheckResult.NoResponse

        monitor.checkNow()
        runCurrent()

        assertEquals(ServerConnectionState.Offline, monitor.state.value)
    }

    @Test
    fun `parar cancela as verificacoes periodicas`() = runTest {
        val monitor = monitor()
        monitor.start()
        runCurrent()

        monitor.stop()
        advanceTimeBy(300_000)
        advanceUntilIdle()

        assertEquals(1, checkCount)
    }
}
