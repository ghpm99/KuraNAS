package com.kuranas.android.feature.connection.ui

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.R
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.discovery.ServerSearch
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ConnectionUiState(
    val isDiscovering: Boolean = false,
    val discovered: List<DiscoveredServer> = emptyList(),
    val hasSearchFinishedEmpty: Boolean = false,
    val manualUrl: String = "",
    val error: String? = null,
)

@HiltViewModel
class ConnectionViewModel @Inject constructor(
    private val serverSearch: ServerSearch,
    private val serverStore: ServerStore,
    @ApplicationContext private val context: Context,
) : ViewModel() {

    private val _state = MutableStateFlow(ConnectionUiState())
    val state: StateFlow<ConnectionUiState> = _state.asStateFlow()

    private var discoveryJob: Job? = null

    init {
        startDiscovery()
    }

    fun startDiscovery() {
        discoveryJob?.cancel()
        _state.update {
            it.copy(isDiscovering = true, discovered = emptyList(), hasSearchFinishedEmpty = false, error = null)
        }
        discoveryJob = viewModelScope.launch {
            try {
                serverSearch.search().collect { server ->
                    _state.update { it.copy(discovered = it.discovered + server) }
                }
                finishDiscovery()
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (_: Exception) {
                _state.update { it.copy(error = context.getString(R.string.connection_discovery_failed)) }
                finishDiscovery()
            }
        }
    }

    private fun finishDiscovery() {
        _state.update { it.copy(isDiscovering = false, hasSearchFinishedEmpty = it.discovered.isEmpty()) }
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
            _state.update { it.copy(error = context.getString(R.string.connection_invalid_address)) }
            return
        }
        val normalized = if (url.startsWith("http")) url else "http://$url"
        viewModelScope.launch {
            serverStore.saveManual(normalized)
        }
    }
}
