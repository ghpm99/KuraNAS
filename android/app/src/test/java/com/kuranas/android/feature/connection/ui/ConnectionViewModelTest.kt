package com.kuranas.android.feature.connection.ui

import android.content.Context
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.discovery.ServerSearch
import com.kuranas.android.core.server.ServerStore
import io.mockk.every
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
        return ConnectionViewModel(serverSearch, serverStore, context)
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
        val viewModel = ConnectionViewModel(serverSearch, serverStore, context)

        every { serverSearch.search() } returns flowOf(foundServer)
        viewModel.startDiscovery()
        firstSearchRelease.complete(Unit)

        assertEquals(listOf(foundServer), viewModel.state.value.discovered)
        assertFalse(viewModel.state.value.isDiscovering)
    }
}
