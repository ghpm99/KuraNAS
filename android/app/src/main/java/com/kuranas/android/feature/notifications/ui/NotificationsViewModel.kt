package com.kuranas.android.feature.notifications.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.feature.notifications.data.NotificationDto
import com.kuranas.android.feature.notifications.data.NotificationsApi
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

data class NotificationsUiState(
    val isLoading: Boolean = true,
    val notifications: List<NotificationDto> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class NotificationsViewModel @Inject constructor(private val api: NotificationsApi) : ViewModel() {

    private val _state = MutableStateFlow(NotificationsUiState())
    val state: StateFlow<NotificationsUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true) }
            when (val r = safeApiCall { api.listNotifications() }) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, notifications = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun markAsRead(id: Int) {
        viewModelScope.launch {
            safeApiCall { api.markAsRead(id) }
            _state.update { state ->
                state.copy(notifications = state.notifications.map { if (it.id == id) it.copy(isRead = true) else it })
            }
        }
    }

    fun markAllAsRead() {
        viewModelScope.launch {
            safeApiCall { api.markAllAsRead() }
            _state.update { state ->
                state.copy(notifications = state.notifications.map { it.copy(isRead = true) })
            }
        }
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object NotificationsModule {
    @Provides
    @Singleton
    fun provideNotificationsApi(retrofit: Retrofit): NotificationsApi = retrofit.create(NotificationsApi::class.java)
}
