package com.kuranas.android.feature.connection.ui

data class ServerSearchPanel(
    val isSearchingNoticeVisible: Boolean,
    val isNoServerNoticeVisible: Boolean,
    val isSearchAgainEnabled: Boolean,
)

data class ManualAddressPanel(
    val isConnectEnabled: Boolean,
    val isProgressVisible: Boolean,
    val isErrorVisible: Boolean,
)

fun serverSearchPanelOf(state: ConnectionUiState): ServerSearchPanel = ServerSearchPanel(
    isSearchingNoticeVisible = state.isDiscovering,
    isNoServerNoticeVisible = state.hasSearchFinishedEmpty && !state.isDiscovering,
    isSearchAgainEnabled = !state.isDiscovering,
)

fun manualAddressPanelOf(state: ConnectionUiState): ManualAddressPanel = ManualAddressPanel(
    isConnectEnabled = !state.isCheckingManualAddress,
    isProgressVisible = state.isCheckingManualAddress,
    isErrorVisible = state.error != null,
)
