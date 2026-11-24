package com.kuranas.android.navigation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.server.ChangeServerAction
import com.kuranas.android.core.server.ServerConnectionMonitor
import com.kuranas.android.core.server.ServerConnectionState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ConnectionBannerViewModel @Inject constructor(
    private val monitor: ServerConnectionMonitor,
    private val changeServerAction: ChangeServerAction,
) : ViewModel() {

    val connectionState: StateFlow<ServerConnectionState> = monitor.state

    val reconnections: Flow<Unit> = monitor.reconnections

    fun retryNow() = monitor.checkNow()

    fun changeServer() {
        viewModelScope.launch { changeServerAction.changeServer() }
    }
}
