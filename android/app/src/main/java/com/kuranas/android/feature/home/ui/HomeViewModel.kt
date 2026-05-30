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
import kotlinx.coroutines.flow.update
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

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.value = HomeUiState.Loading
            val statsResult = repository.getStats()
            val recentResult = repository.getRecentFiles()
            _state.value = when {
                statsResult is AppResult.Success && recentResult is AppResult.Success ->
                    HomeUiState.Success(statsResult.data, recentResult.data)
                statsResult is AppResult.Error -> HomeUiState.Error(statsResult.message)
                else -> HomeUiState.Error("Falha ao carregar dados")
            }
        }
    }
}
