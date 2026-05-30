package com.kuranas.android.navigation

import androidx.lifecycle.ViewModel
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.server.ServerStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.stateIn

@HiltViewModel
class RootViewModel @Inject constructor(serverStore: ServerStore) : ViewModel() {

    val serverState: StateFlow<ServerState> = serverStore.serverState
        .stateIn(viewModelScope, SharingStarted.Eagerly, ServerState.NotConfigured)
}
