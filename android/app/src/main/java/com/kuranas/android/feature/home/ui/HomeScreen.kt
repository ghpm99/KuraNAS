package com.kuranas.android.feature.home.ui

import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.MusicNote
import androidx.compose.material.icons.filled.Photo
import androidx.compose.material.icons.filled.VideoLibrary
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.FileSizeText
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass

@Composable
fun HomeScreen(
    onOpenMenu: () -> Unit,
    onOpenFiles: () -> Unit,
    onOpenMusic: () -> Unit,
    onOpenVideo: () -> Unit,
    onOpenImages: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = "Início",
            leadingIcon = Icons.Default.Menu,
            onLeadingClick = onOpenMenu,
        )
        when (val s = state) {
            is HomeUiState.Loading -> LoadingView()
            is HomeUiState.Error -> ErrorView(s.message, onRetry = viewModel::load)
            is HomeUiState.Success -> {
                LazyColumn(
                    contentPadding = PaddingValues(bottom = 24.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    item {
                        StatsCard(
                            spaceUsed = s.stats.totalSpaceUsed,
                            totalFiles = s.stats.totalFiles,
                        )
                    }
                    item {
                        Text("Acesso rápido", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(vertical = 4.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            QuickAccessItem("Arquivos", Icons.Default.Folder, Modifier.weight(1f), onOpenFiles)
                            QuickAccessItem("Música", Icons.Default.MusicNote, Modifier.weight(1f), onOpenMusic)
                            QuickAccessItem("Vídeos", Icons.Default.VideoLibrary, Modifier.weight(1f), onOpenVideo)
                            QuickAccessItem("Imagens", Icons.Default.Photo, Modifier.weight(1f), onOpenImages)
                        }
                    }
                    if (s.recentFiles.isNotEmpty()) {
                        item {
                            Spacer(Modifier.height(4.dp))
                            Text("Acessados recentemente", style = MaterialTheme.typography.titleMedium)
                        }
                        items(s.recentFiles.take(10)) { file ->
                            RecentFileItem(name = file.name, size = file.size, mimeType = file.mimeType)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun StatsCard(spaceUsed: Long, totalFiles: Long) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Strong)
            .padding(20.dp),
    ) {
        Text("Armazenamento", style = MaterialTheme.typography.labelMedium)
        Spacer(Modifier.height(4.dp))
        FileSizeText(spaceUsed, style = MaterialTheme.typography.headlineMedium)
        Text("$totalFiles arquivos", style = MaterialTheme.typography.bodySmall)
    }
}

@Composable
private fun QuickAccessItem(label: String, icon: ImageVector, modifier: Modifier, onClick: () -> Unit) {
    Column(
        modifier = modifier
            .glass(GlassLevel.Light, radius = 16.dp)
            .clickable(onClick = onClick)
            .padding(12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(icon, contentDescription = label, modifier = Modifier.size(28.dp), tint = MaterialTheme.colorScheme.primary)
        Spacer(Modifier.height(4.dp))
        Text(label, style = MaterialTheme.typography.labelMedium)
    }
}

@Composable
private fun RecentFileItem(name: String, size: Long, mimeType: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Flat, radius = 12.dp)
            .padding(horizontal = 16.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(name, style = MaterialTheme.typography.bodyMedium, maxLines = 1)
            Text(mimeType, style = MaterialTheme.typography.bodySmall, maxLines = 1)
        }
        FileSizeText(size)
    }
}
