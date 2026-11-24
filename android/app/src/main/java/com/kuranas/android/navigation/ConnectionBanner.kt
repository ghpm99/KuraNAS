package com.kuranas.android.navigation

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.ChangeServerConfirmationDialog
import com.kuranas.android.ui.theme.StatusAlert

@Composable
fun ConnectionBanner(onRetryNow: () -> Unit, onChangeServer: () -> Unit, modifier: Modifier = Modifier) {
    var isChangeServerDialogVisible by rememberSaveable { mutableStateOf(false) }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(StatusAlert.copy(alpha = 0.16f))
            .padding(horizontal = 16.dp, vertical = 8.dp),
    ) {
        Text(
            text = stringResource(R.string.connection_banner_offline),
            style = MaterialTheme.typography.bodyMedium,
            color = StatusAlert,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            TextButton(onClick = onRetryNow) { Text(stringResource(R.string.connection_banner_retry_now)) }
            TextButton(onClick = { isChangeServerDialogVisible = true }) {
                Text(stringResource(R.string.action_change_server))
            }
        }
    }

    if (isChangeServerDialogVisible) {
        ChangeServerConfirmationDialog(
            onConfirm = {
                isChangeServerDialogVisible = false
                onChangeServer()
            },
            onDismiss = { isChangeServerDialogVisible = false },
        )
    }
}
