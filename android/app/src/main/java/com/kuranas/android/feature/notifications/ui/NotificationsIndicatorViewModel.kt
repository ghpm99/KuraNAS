package com.kuranas.android.feature.notifications.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.feature.notifications.data.NotificationsApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class NotificationsIndicatorViewModel @Inject constructor(
    private val api: NotificationsApi,
) : ViewModel() {

    private val _unreadCount = MutableStateFlow(0)
    val unreadCount: StateFlow<Int> = _unreadCount.asStateFlow()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            _unreadCount.value = when (val outcome = safeApiCall { api.getUnreadCount() }) {
                is AppResult.Success -> outcome.data.count
                is AppResult.Error -> 0
            }
        }
    }
}
