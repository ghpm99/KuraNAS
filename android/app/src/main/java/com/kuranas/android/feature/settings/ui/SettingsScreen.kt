package com.kuranas.android.feature.settings.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Circle
import androidx.compose.material.icons.filled.Cloud
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Logout
import androidx.compose.material.icons.filled.Update
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.ui.theme.StatusNegative
import com.kuranas.android.ui.theme.StatusPositive

@Composable
fun SettingsScreen(viewModel: SettingsViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var showForgetDialog by remember { mutableStateOf(false) }

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = stringResource(R.string.settings_title))

        if (state.isLoading) {
            LoadingView()
            return@Column
        }

        LazyColumn(contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            item {
                Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light).padding(16.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Icon(Icons.Default.Cloud, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
                        Text(stringResource(R.string.settings_server), style = MaterialTheme.typography.titleMedium)
                    }
                    Spacer(Modifier.height(8.dp))
                    Text(state.currentServerUrl, style = MaterialTheme.typography.bodyMedium)
                    Spacer(Modifier.height(4.dp))
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        Icon(Icons.Default.Circle, contentDescription = null, modifier = Modifier.size(8.dp), tint = StatusPositive)
                        Text(stringResource(R.string.settings_connected), style = MaterialTheme.typography.bodySmall, color = StatusPositive)
                    }
                }
            }

            state.about?.let { about ->
                item {
                    Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Flat).padding(16.dp)) {
                        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            Icon(Icons.Default.Info, contentDescription = null, tint = MaterialTheme.colorScheme.secondary)
                            Text(stringResource(R.string.settings_about_server), style = MaterialTheme.typography.titleMedium)
                        }
                        Spacer(Modifier.height(8.dp))
                        Text(stringResource(R.string.settings_version, about.version), style = MaterialTheme.typography.bodyMedium)
                        Text(stringResource(R.string.settings_platform, about.platform), style = MaterialTheme.typography.bodySmall)
                        Text(stringResource(R.string.settings_uptime, about.uptime), style = MaterialTheme.typography.bodySmall)
                    }
                }
            }

            state.updateStatus?.let { update ->
                if (update.hasUpdate) {
                    item {
                        Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Strong).padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                Icon(Icons.Default.Update, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
                                Text(stringResource(R.string.settings_update_available), style = MaterialTheme.typography.titleMedium)
                            }
                            Spacer(Modifier.height(4.dp))
                            Text(stringResource(R.string.settings_update_detail, update.latestVersion.orEmpty(), update.currentVersion.orEmpty()), style = MaterialTheme.typography.bodySmall)
                        }
                    }
                }
            }

            item {
                Button(
                    onClick = { showForgetDialog = true },
                    colors = ButtonDefaults.buttonColors(containerColor = StatusNegative),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Icon(Icons.Default.Logout, contentDescription = null, modifier = Modifier.size(18.dp))
                    Spacer(Modifier.size(8.dp))
                    Text(stringResource(R.string.action_forget_server))
                }
            }
        }
    }

    if (showForgetDialog) {
        AlertDialog(
            onDismissRequest = { showForgetDialog = false },
            title = { Text(stringResource(R.string.settings_forget_dialog_title)) },
            text = { Text(stringResource(R.string.settings_forget_dialog_text)) },
            confirmButton = {
                TextButton(onClick = { viewModel.forgetServer(); showForgetDialog = false }) { Text(stringResource(R.string.action_forget)) }
            },
            dismissButton = { TextButton(onClick = { showForgetDialog = false }) { Text(stringResource(R.string.action_cancel)) } },
        )
    }
}
