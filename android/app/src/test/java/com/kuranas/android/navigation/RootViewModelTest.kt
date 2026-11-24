package com.kuranas.android.navigation

import com.kuranas.android.core.server.AccessDenial
import com.kuranas.android.core.server.AccessDeniedEvents
import com.kuranas.android.core.server.ChangeServerAction
import com.kuranas.android.core.server.ReconnectionOutcome
import com.kuranas.android.core.server.ServerReconnection
import com.kuranas.android.core.server.ServerSelectionEvents
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
    private val serverSelectionEvents = ServerSelectionEvents()
    private val accessDeniedEvents = AccessDeniedEvents()
    private val changeServerAction = mockk<ChangeServerAction>(relaxed = true)
    private val denial = AccessDenial("10.0.0.9", "blocked")
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

    private fun createViewModel() = RootViewModel(serverStore, serverReconnection, serverSelectionEvents, accessDeniedEvents, changeServerAction)

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
    fun `Unreachable abre a tela de conexao`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Unreachable)
        advanceUntilIdle()

        assertEquals(RootDestination.Connection, viewModel.destination.value)
    }

    @Test
    fun `AccessDenied na abertura mostra a tela de acesso negado com o ip`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.AccessDenied(denial))
        advanceUntilIdle()

        assertEquals(RootDestination.AccessDenied(denial), viewModel.destination.value)
    }

    @Test
    fun `recusa no meio da sessao troca o app pela tela de acesso negado`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Reachable)
        advanceUntilIdle()
        assertEquals(RootDestination.App, viewModel.destination.value)

        accessDeniedEvents.report(denial)
        advanceUntilIdle()

        assertEquals(RootDestination.AccessDenied(denial), viewModel.destination.value)
    }

    @Test
    fun `tentar de novo com sucesso limpa a recusa e volta ao app`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.AccessDenied(denial))
        advanceUntilIdle()
        coEvery { serverReconnection.ensureReachable() } returns ReconnectionOutcome.Reachable

        viewModel.retryAfterAccessDenied()
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
        assertEquals(null, accessDeniedEvents.denial.value)
    }

    @Test
    fun `tentar de novo ainda recusado continua na tela de acesso negado`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.AccessDenied(denial))
        advanceUntilIdle()
        val newDenial = AccessDenial("10.0.0.10", null)
        coEvery { serverReconnection.ensureReachable() } returns ReconnectionOutcome.AccessDenied(newDenial)

        viewModel.retryAfterAccessDenied()
        advanceUntilIdle()

        assertEquals(RootDestination.AccessDenied(newDenial), viewModel.destination.value)
    }

    @Test
    fun `trocar servidor na tela de acesso negado limpa a recusa e abre a conexao`() = runTest {
        coEvery { changeServerAction.changeServer() } coAnswers { serverStates.value = ServerState.NotConfigured }
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.AccessDenied(denial))
        advanceUntilIdle()

        viewModel.changeServerAfterAccessDenied()
        advanceUntilIdle()

        assertEquals(RootDestination.Connection, viewModel.destination.value)
        assertEquals(null, accessDeniedEvents.denial.value)
    }

    @Test
    fun `escolher outro servidor depois de Unreachable entra no app`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Unreachable)
        advanceUntilIdle()
        serverStates.value = ServerState.Manual("http://10.0.0.5:8000")
        serverSelectionEvents.notifyServerSelected()
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
    }

    @Test
    fun `escolher o mesmo servidor depois de Unreachable entra no app`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Unreachable)
        advanceUntilIdle()
        serverSelectionEvents.notifyServerSelected()
        advanceUntilIdle()

        assertEquals(RootDestination.App, viewModel.destination.value)
    }

    @Test
    fun `sem escolha depois de Unreachable continua na tela de conexao`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Unreachable)
        advanceUntilIdle()
        serverStates.value = ServerState.Manual("http://10.0.0.5:8000")
        advanceUntilIdle()

        assertEquals(RootDestination.Connection, viewModel.destination.value)
    }

    @Test
    fun `trocar servidor depois de Reachable volta para a tela de conexao e permite escolher de novo`() = runTest {
        val viewModel = createViewModel()
        pendingOutcome.complete(ReconnectionOutcome.Reachable)
        advanceUntilIdle()

        serverStates.value = ServerState.NotConfigured
        advanceUntilIdle()
        assertEquals(RootDestination.Connection, viewModel.destination.value)

        serverStates.value = ServerState.Manual("http://10.0.0.5:8000")
        serverSelectionEvents.notifyServerSelected()
        advanceUntilIdle()
        assertEquals(RootDestination.App, viewModel.destination.value)
    }
}
