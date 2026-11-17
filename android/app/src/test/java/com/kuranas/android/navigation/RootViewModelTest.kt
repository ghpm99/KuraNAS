package com.kuranas.android.navigation

import com.kuranas.android.core.server.ReconnectionOutcome
import com.kuranas.android.core.server.ServerReconnection
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.server.ServerStore
import io.mockk.coEvery
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class RootViewModelTest {

    private val rememberedServer = ServerState.Found("http://192.168.0.10:8000", "nas")
    private val serverStates = MutableStateFlow<ServerState>(rememberedServer)
    private val serverStore = mockk<ServerStore>()
    private val serverReconnection = mockk<ServerReconnection>()
    private val pendingOutcome = CompletableDeferred<ReconnectionOutcome>()

    @Before
    fun setUp() {
        Dispatchers.setMain(StandardTestDispatcher())
        every { serverStore.serverState } returns serverStates
        coEvery { serverReconnection.ensureReachable() } coAnswers { pendingOutcome.await() }
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun createViewModel() = RootViewModel(serverStore, serverReconnection)

    @Test
    fun `fica verificando ate o desfecho chegar`() = runTest {
        val viewModel = createViewModel()
        advanceUntilIdle()

        assertEquals(RootDestination.Checking, viewModel.destination.value)
    }

    @Test
    fun `Reachable abre o app`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Reachable)
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
    }

    @Test
    fun `Relocated abre o app com o servidor novo`() = runTest {
        val viewModel = createViewModel()
        serverStates.value = ServerState.Found("http://192.168.0.20:8000", "nas")
        pendingOutcome.complete(ReconnectionOutcome.Relocated)
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
    }

    @Test
    fun `NoServer abre a tela de conexao`() = runTest {
        serverStates.value = ServerState.NotConfigured
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.NoServer)
        advanceUntilIdle()

        assertEquals(RootDestination.Connection, viewModel.destination.value)
    }

    @Test
    fun `Unreachable e AccessDenied abrem a tela de conexao`() = runTest {
        listOf(ReconnectionOutcome.Unreachable, ReconnectionOutcome.AccessDenied).forEach { outcome ->
            val deferredOutcome = CompletableDeferred(outcome)
            coEvery { serverReconnection.ensureReachable() } coAnswers { deferredOutcome.await() }
            val viewModel = createViewModel()
            advanceUntilIdle()

            assertEquals(outcome.toString(), RootDestination.Connection, viewModel.destination.value)
        }
    }

    @Test
    fun `escolher outro servidor depois de Unreachable entra no app`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Unreachable)
        advanceUntilIdle()
        serverStates.value = ServerState.Manual("http://10.0.0.5:8000")
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
    }
}
