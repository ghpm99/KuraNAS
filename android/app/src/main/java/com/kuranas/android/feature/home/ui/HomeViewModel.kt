package com.kuranas.android.feature.home.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.home.data.HomeRepository
import com.kuranas.android.feature.home.data.HomeStats
import com.kuranas.android.feature.home.data.RecentFileDto
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed class HomeUiState {
    object Loading : HomeUiState()
    data class Success(val stats: HomeStats, val recentFiles: List<RecentFileDto>) : HomeUiState()
    data class Error(val message: String) : HomeUiState()
}

@HiltViewModel
class HomeViewModel @Inject constructor(private val repository: HomeRepository) : ViewModel() {

    private val _state = MutableStateFlow<HomeUiState>(HomeUiState.Loading)
    val state: StateFlow<HomeUiState> = _state.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.value = HomeUiState.Loading
            _state.value = when (val result = repository.getHomeData()) {
                is AppResult.Success -> HomeUiState.Success(result.data.stats, result.data.recentFiles)
                is AppResult.Error -> HomeUiState.Error(result.message)
            }
        }
    }

    /**
     * Recarrega sem voltar ao estado de Loading (mantém o conteúdo atual na tela
     * enquanto busca). Usado pelo pull-to-refresh e pelo refetch ao retomar a tela.
     */
    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            when (val result = repository.getHomeData()) {
                is AppResult.Success ->
                    _state.value = HomeUiState.Success(result.data.stats, result.data.recentFiles)
                is AppResult.Error ->
                    if (_state.value !is HomeUiState.Success) {
                        _state.value = HomeUiState.Error(result.message)
                    }
            }
            _isRefreshing.value = false
        }
    }
}
