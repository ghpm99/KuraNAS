package com.kuranas.android.feature.connection.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.discovery.NsdDiscovery
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ConnectionUiState(
    val isDiscovering: Boolean = false,
    val discovered: List<DiscoveredServer> = emptyList(),
    val manualUrl: String = "",
    val error: String? = null,
)

@HiltViewModel
class ConnectionViewModel @Inject constructor(
    private val nsdDiscovery: NsdDiscovery,
    private val serverStore: ServerStore,
) : ViewModel() {

    private val _state = MutableStateFlow(ConnectionUiState())
    val state: StateFlow<ConnectionUiState> = _state.asStateFlow()

    init {
        startDiscovery()
    }

    fun startDiscovery() {
        _state.update { it.copy(isDiscovering = true, discovered = emptyList(), error = null) }
        viewModelScope.launch {
            try {
                nsdDiscovery.discover().collect { server ->
                    _state.update { it.copy(discovered = it.discovered + server) }
                }
            } catch (e: Exception) {
                _state.update { it.copy(error = "Falha na descoberta automática", isDiscovering = false) }
            }
        }
    }

    fun onManualUrlChange(url: String) {
        _state.update { it.copy(manualUrl = url, error = null) }
    }

    fun connectToDiscovered(server: DiscoveredServer) {
        viewModelScope.launch {
            serverStore.saveDiscovered(server.url, server.name)
        }
    }

    fun connectManual() {
        val url = _state.value.manualUrl.trim()
        if (url.isEmpty()) {
            _state.update { it.copy(error = "Insira um endereço válido") }
            return
        }
        val normalized = if (url.startsWith("http")) url else "http://$url"
        viewModelScope.launch {
            serverStore.saveManual(normalized)
        }
    }
}
