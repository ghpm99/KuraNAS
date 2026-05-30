package com.kuranas.android.feature.settings.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.server.ServerStore
import com.kuranas.android.feature.settings.data.AboutDto
import com.kuranas.android.feature.settings.data.ConfigApi
import com.kuranas.android.feature.settings.data.ServerSettingsDto
import com.kuranas.android.feature.settings.data.UpdateStatusDto
import dagger.Module
import dagger.Provides
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

data class SettingsUiState(
    val isLoading: Boolean = true,
    val about: AboutDto? = null,
    val settings: ServerSettingsDto? = null,
    val updateStatus: UpdateStatusDto? = null,
    val currentServerUrl: String = "",
    val error: String? = null,
)

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val api: ConfigApi,
    private val serverStore: ServerStore,
) : ViewModel() {

    private val _state = MutableStateFlow(SettingsUiState())
    val state: StateFlow<SettingsUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true) }
            val about = safeApiCall { api.getAbout() }
            val settings = safeApiCall { api.getSettings() }
            val updateStatus = safeApiCall { api.getUpdateStatus() }
            serverStore.serverUrl.collect { url ->
                _state.update {
                    it.copy(
                        isLoading = false,
                        about = (about as? AppResult.Success)?.data,
                        settings = (settings as? AppResult.Success)?.data,
                        updateStatus = (updateStatus as? AppResult.Success)?.data,
                        currentServerUrl = url ?: "",
                        error = (about as? AppResult.Error)?.message,
                    )
                }
                return@collect
            }
        }
    }

    fun forgetServer() {
        viewModelScope.launch { serverStore.clear() }
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object ConfigModule {
    @Provides
    @Singleton
    fun provideConfigApi(retrofit: Retrofit): ConfigApi = retrofit.create(ConfigApi::class.java)
}
