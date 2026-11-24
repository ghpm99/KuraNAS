package com.kuranas.android.feature.connection.ui

import android.content.Context
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.discovery.ServerSearch
import com.kuranas.android.R
import com.kuranas.android.core.server.AccessDenial
import com.kuranas.android.core.server.ServerCheckResult
import com.kuranas.android.core.server.ServerHealthChecker
import com.kuranas.android.core.server.ServerSelectionEvents
import com.kuranas.android.core.server.ServerStore
import io.mockk.every
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.emptyFlow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ConnectionViewModelTest {

    private val serverSearch = mockk<ServerSearch>()
    private val serverStore = mockk<ServerStore>(relaxed = true)
    private val serverHealthChecker = mockk<ServerHealthChecker>()
    private val serverSelectionEvents = ServerSelectionEvents()
    private val context = mockk<Context>(relaxed = true)
    private val foundServer = DiscoveredServer("nas", "192.168.0.10", 8000)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun viewModelSearching(servers: Flow<DiscoveredServer>): ConnectionViewModel {
        every { serverSearch.search() } returns servers
        return ConnectionViewModel(serverSearch, serverStore, serverHealthChecker, serverSelectionEvents, context)
    }

    @Test
    fun `stops discovering and keeps found servers when search ends`() {
        val viewModel = viewModelSearching(flowOf(foundServer))

        assertFalse(viewModel.state.value.isDiscovering)
        assertEquals(listOf(foundServer), viewModel.state.value.discovered)
        assertFalse(viewModel.state.value.hasSearchFinishedEmpty)
    }

    @Test
    fun `flags empty result when search ends without servers`() {
        val viewModel = viewModelSearching(emptyFlow())

        assertFalse(viewModel.state.value.isDiscovering)
        assertTrue(viewModel.state.value.hasSearchFinishedEmpty)
    }

    @Test
    fun `stops discovering when search fails`() {
        val viewModel = viewModelSearching(flow { throw IllegalStateException("boom") })

        assertFalse(viewModel.state.value.isDiscovering)
        assertTrue(viewModel.state.value.hasSearchFinishedEmpty)
    }

    @Test
    fun `second search cancels the first`() {
        val firstSearchRelease = CompletableDeferred<Unit>()
        val firstSearch = flow<DiscoveredServer> {
            firstSearchRelease.await()
            emit(DiscoveredServer("stale", "10.0.0.1", 8000))
        }
        every { serverSearch.search() } returns firstSearch
        val viewModel = ConnectionViewModel(serverSearch, serverStore, serverHealthChecker, serverSelectionEvents, context)

        every { serverSearch.search() } returns flowOf(foundServer)
        viewModel.startDiscovery()
        firstSearchRelease.complete(Unit)

        assertEquals(listOf(foundServer), viewModel.state.value.discovered)
        assertFalse(viewModel.state.value.isDiscovering)
    }

    private fun viewModelTyping(typedAddress: String): ConnectionViewModel {
        val viewModel = viewModelSearching(emptyFlow())
        viewModel.onManualUrlChange(typedAddress)
        return viewModel
    }

    @Test
    fun `saves normalized address when server is confirmed`() {
        coEvery { serverHealthChecker.check("http://192.168.0.10:8000") } returns ServerCheckResult.Confirmed
        val viewModel = viewModelTyping("192.168.0.10")

        viewModel.connectManual()

        coVerify { serverStore.saveManual("http://192.168.0.10:8000") }
        assertFalse(viewModel.state.value.isCheckingManualAddress)
        assertEquals(null, viewModel.state.value.error)
    }

    @Test
    fun `confirmed manual address signals the selection`() {
        coEvery { serverHealthChecker.check("http://192.168.0.10:8000") } returns ServerCheckResult.Confirmed
        val viewModel = viewModelTyping("192.168.0.10")

        viewModel.connectManual()

        assertEquals(1, serverSelectionEvents.selectionCount.value)
    }

    @Test
    fun `connecting to a discovered server saves it and signals the selection`() {
        val viewModel = viewModelSearching(flowOf(foundServer))

        viewModel.connectToDiscovered(foundServer)

        coVerify { serverStore.saveDiscovered(foundServer.url, foundServer.name) }
        assertEquals(1, serverSelectionEvents.selectionCount.value)
    }

    @Test
    fun `rejected check does not save and shows its reason`() {
        every { context.getString(R.string.connection_no_response) } returns "no response"
        every { context.getString(R.string.connection_not_kuranas) } returns "not kuranas"
        every { context.getString(R.string.connection_access_denied) } returns "access denied"
        every { context.getString(R.string.connection_access_denied_with_ip, "10.0.0.9") } returns "access denied for 10.0.0.9"
        val expectedErrors = mapOf(
            ServerCheckResult.NoResponse to "no response",
            ServerCheckResult.NotKuraNas to "not kuranas",
            ServerCheckResult.AccessDenied(AccessDenial("10.0.0.9", null)) to "access denied for 10.0.0.9",
            ServerCheckResult.AccessDenied(AccessDenial(null, null)) to "access denied",
        )

        expectedErrors.forEach { (checkResult, expectedError) ->
            coEvery { serverHealthChecker.check(any()) } returns checkResult
            val viewModel = viewModelTyping("192.168.0.10")

            viewModel.connectManual()

            assertEquals(expectedError, viewModel.state.value.error)
            assertFalse(viewModel.state.value.isCheckingManualAddress)
        }
        coVerify(exactly = 0) { serverStore.saveManual(any()) }
        assertEquals(0, serverSelectionEvents.selectionCount.value)
    }

    @Test
    fun `invalid address does not call the checker`() {
        every { context.getString(R.string.connection_invalid_address) } returns "invalid"
        val viewModel = viewModelTyping("ftp://nas")

        viewModel.connectManual()

        assertEquals("invalid", viewModel.state.value.error)
        coVerify(exactly = 0) { serverHealthChecker.check(any()) }
    }
}
