package com.kuranas.android.feature.connection.ui

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Dns
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Router
import androidx.compose.material.icons.filled.Storage
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.glass

@Composable
fun ConnectionScreen(viewModel: ConnectionViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val searchPanel = serverSearchPanelOf(state)
    val manualPanel = manualAddressPanelOf(state)
    val infiniteTransition = rememberInfiniteTransition(label = "scan")
    val rotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(tween(2000, easing = LinearEasing), RepeatMode.Restart),
        label = "rotation",
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp, vertical = 48.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Icon(
            imageVector = Icons.Default.Storage,
            contentDescription = null,
            modifier = Modifier.size(72.dp),
            tint = MaterialTheme.colorScheme.primary,
        )
        Spacer(Modifier.height(16.dp))
        Text(stringResource(R.string.app_name), style = MaterialTheme.typography.headlineLarge)
        Text(stringResource(R.string.connection_connect_to_server), style = MaterialTheme.typography.bodyMedium)
        Spacer(Modifier.height(32.dp))

        Text(stringResource(R.string.connection_servers_on_network), style = MaterialTheme.typography.titleMedium)

        if (searchPanel.isSearchingNoticeVisible) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
            ) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp).rotate(rotation),
                    strokeWidth = 2.dp,
                )
                Text(stringResource(R.string.connection_searching_server), style = MaterialTheme.typography.bodyMedium)
            }
        }

        if (searchPanel.isNoServerNoticeVisible) {
            Text(
                stringResource(R.string.connection_no_servers_found),
                style = MaterialTheme.typography.titleSmall,
                modifier = Modifier.padding(top = 8.dp),
            )
            Text(
                stringResource(R.string.connection_check_same_wifi),
                style = MaterialTheme.typography.bodySmall,
                modifier = Modifier.padding(vertical = 4.dp),
            )
        }

        TextButton(onClick = viewModel::startDiscovery, enabled = searchPanel.isSearchAgainEnabled) {
            Icon(Icons.Default.Refresh, contentDescription = null, modifier = Modifier.size(18.dp))
            Spacer(Modifier.size(8.dp))
            Text(stringResource(R.string.connection_search_again))
        }

        LazyColumn(modifier = Modifier.weight(1f, fill = false)) {
            items(state.discovered) { server ->
                DiscoveredServerItem(server = server, onClick = { viewModel.connectToDiscovered(server) })
            }
        }

        Spacer(Modifier.height(24.dp))
        Text(stringResource(R.string.connection_or_connect_manually), style = MaterialTheme.typography.labelMedium)
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = state.manualUrl,
            onValueChange = viewModel::onManualUrlChange,
            label = { Text(stringResource(R.string.connection_server_address_label)) },
            placeholder = { Text(stringResource(R.string.connection_server_address_placeholder)) },
            singleLine = true,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Uri,
                imeAction = ImeAction.Done,
            ),
            keyboardActions = KeyboardActions(onDone = { viewModel.connectManual() }),
            isError = manualPanel.isErrorVisible,
            supportingText = state.error?.let { { Text(it) } },
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(Modifier.height(12.dp))
        Button(
            onClick = viewModel::connectManual,
            enabled = manualPanel.isConnectEnabled,
            modifier = Modifier.fillMaxWidth(),
        ) {
            if (manualPanel.isProgressVisible) {
                CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
            } else {
                Icon(Icons.Default.Dns, contentDescription = null, modifier = Modifier.size(18.dp))
            }
            Spacer(Modifier.size(8.dp))
            Text(stringResource(R.string.action_connect))
        }
    }
}

@Composable
private fun DiscoveredServerItem(server: DiscoveredServer, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp)
            .glass(GlassLevel.Light, radius = 12.dp)
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(Icons.Default.Router, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
        Column {
            Text(server.name, style = MaterialTheme.typography.titleMedium)
            Text(server.url, style = MaterialTheme.typography.bodySmall)
        }
    }
}
