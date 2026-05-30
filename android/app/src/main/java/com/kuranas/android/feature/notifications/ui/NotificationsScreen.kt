package com.kuranas.android.feature.notifications.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.DoneAll
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.NotificationsNone
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.notifications.data.NotificationDto

@Composable
fun NotificationsScreen(viewModel: NotificationsViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = "Notificações",
            trailingIcon = Icons.Default.DoneAll,
            onTrailingClick = viewModel::markAllAsRead,
        )
        when {
            state.isLoading -> LoadingView()
            state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
            state.notifications.isEmpty() -> EmptyView("Nenhuma notificação", icon = Icons.Default.NotificationsNone)
            else -> LazyColumn(contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                items(state.notifications, key = { it.id }) { notification ->
                    NotificationItem(notification = notification, onRead = { viewModel.markAsRead(notification.id) })
                }
            }
        }
    }
}

@Composable
private fun NotificationItem(notification: NotificationDto, onRead: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(if (notification.isRead) GlassLevel.Flat else GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onRead)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(
            imageVector = if (notification.isRead) Icons.Default.NotificationsNone else Icons.Default.Notifications,
            contentDescription = null,
            modifier = Modifier.size(20.dp),
            tint = if (notification.isRead) MaterialTheme.colorScheme.onSurfaceVariant else MaterialTheme.colorScheme.primary,
        )
        Column(modifier = Modifier.weight(1f)) {
            Text(notification.title, style = MaterialTheme.typography.bodyMedium)
            Text(notification.message, style = MaterialTheme.typography.bodySmall, maxLines = 2)
            Text(notification.createdAt, style = MaterialTheme.typography.bodySmall)
        }
    }
}
