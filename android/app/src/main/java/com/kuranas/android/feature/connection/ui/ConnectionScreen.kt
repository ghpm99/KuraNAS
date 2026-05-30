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
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.discovery.DiscoveredServer
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.glass

@Composable
fun ConnectionScreen(viewModel: ConnectionViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
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
        Text("KuraNAS", style = MaterialTheme.typography.headlineLarge)
        Text("Conectar ao servidor", style = MaterialTheme.typography.bodyMedium)
        Spacer(Modifier.height(32.dp))

        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Servidores na rede", style = MaterialTheme.typography.titleMedium)
            if (state.isDiscovering) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp).rotate(rotation),
                    strokeWidth = 2.dp,
                )
            } else {
                IconButton(onClick = viewModel::startDiscovery) {
                    Icon(Icons.Default.Refresh, contentDescription = "Buscar novamente")
                }
            }
        }

        if (state.discovered.isEmpty() && !state.isDiscovering) {
            Text(
                "Nenhum servidor encontrado na rede local",
                style = MaterialTheme.typography.bodySmall,
                modifier = Modifier.padding(vertical = 8.dp),
            )
        }

        LazyColumn(modifier = Modifier.weight(1f, fill = false)) {
            items(state.discovered) { server ->
                DiscoveredServerItem(server = server, onClick = { viewModel.connectToDiscovered(server) })
            }
        }

        Spacer(Modifier.height(24.dp))
        Text("Ou conectar manualmente", style = MaterialTheme.typography.labelMedium)
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = state.manualUrl,
            onValueChange = viewModel::onManualUrlChange,
            label = { Text("Endereço do servidor") },
            placeholder = { Text("192.168.1.100:8000") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Uri,
                imeAction = ImeAction.Done,
            ),
            keyboardActions = KeyboardActions(onDone = { viewModel.connectManual() }),
            isError = state.error != null,
            supportingText = state.error?.let { { Text(it) } },
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(Modifier.height(12.dp))
        Button(onClick = viewModel::connectManual, modifier = Modifier.fillMaxWidth()) {
            Icon(Icons.Default.Dns, contentDescription = null, modifier = Modifier.size(18.dp))
            Spacer(Modifier.size(8.dp))
            Text("Conectar")
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
