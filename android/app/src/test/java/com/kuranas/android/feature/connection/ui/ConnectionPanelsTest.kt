package com.kuranas.android.feature.connection.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class ConnectionPanelsTest {

    @Test
    fun `while searching shows notice and disables search again`() {
        val panel = serverSearchPanelOf(ConnectionUiState(isDiscovering = true))

        assertEquals(ServerSearchPanel(true, false, false), panel)
    }

    @Test
    fun `finished empty shows no server notice and enables search again`() {
        val panel = serverSearchPanelOf(ConnectionUiState(hasSearchFinishedEmpty = true))

        assertEquals(ServerSearchPanel(false, true, true), panel)
    }

    @Test
    fun `no server notice never shows while a new search runs`() {
        val panel = serverSearchPanelOf(ConnectionUiState(isDiscovering = true, hasSearchFinishedEmpty = true))

        assertEquals(false, panel.isNoServerNoticeVisible)
    }

    @Test
    fun `manual address check disables connect and shows progress`() {
        val panel = manualAddressPanelOf(ConnectionUiState(isCheckingManualAddress = true))

        assertEquals(ManualAddressPanel(false, true, false), panel)
    }

    @Test
    fun `manual error is visible`() {
        val panel = manualAddressPanelOf(ConnectionUiState(error = "boom"))

        assertEquals(ManualAddressPanel(true, false, true), panel)
    }
}
