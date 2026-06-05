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
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.FileSizeText
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    onOpenMenu: () -> Unit,
    onOpenFiles: () -> Unit,
    onOpenMusic: () -> Unit,
    onOpenVideo: () -> Unit,
    onOpenImages: () -> Unit,
    onOpenImage: (String) -> Unit,
    onOpenVideoFile: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onOpenFile: (String, String) -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()

    // Recarrega ao retomar a tela (voltar de outra tela / do background),
    // pulando o primeiro ON_RESUME pra não duplicar o load do init.
    var firstResume by rememberSaveable { mutableStateOf(true) }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (firstResume) firstResume = false else viewModel.refresh()
    }

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = stringResource(R.string.nav_home),
            leadingIcon = Icons.Default.Menu,
            onLeadingClick = onOpenMenu,
        )
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
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
                        Text(stringResource(R.string.home_quick_access), style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(vertical = 4.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            QuickAccessItem(stringResource(R.string.nav_files), Icons.Default.Folder, Modifier.weight(1f), onOpenFiles)
                            QuickAccessItem(stringResource(R.string.nav_music), Icons.Default.MusicNote, Modifier.weight(1f), onOpenMusic)
                            QuickAccessItem(stringResource(R.string.nav_videos), Icons.Default.VideoLibrary, Modifier.weight(1f), onOpenVideo)
                            QuickAccessItem(stringResource(R.string.nav_images), Icons.Default.Photo, Modifier.weight(1f), onOpenImages)
                        }
                    }
                    if (s.recentFiles.isNotEmpty()) {
                        item {
                            Spacer(Modifier.height(4.dp))
                            Text(stringResource(R.string.home_recently_accessed), style = MaterialTheme.typography.titleMedium)
                        }
                        items(s.recentFiles.take(10)) { file ->
                            RecentFileItem(
                                name = file.name,
                                size = file.size,
                                mimeType = file.mimeType,
                                onClick = {
                                    when {
                                        file.mimeType.startsWith("image/") -> onOpenImage(file.id.toString())
                                        file.mimeType.startsWith("video/") -> onOpenVideoFile(file.id.toString())
                                        file.mimeType.startsWith("audio/") -> onPlayAudio(file.id.toString())
                                        else -> onOpenFile(file.id.toString(), file.name)
                                    }
                                },
                            )
                        }
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
        Text(stringResource(R.string.home_storage), style = MaterialTheme.typography.labelMedium)
        Spacer(Modifier.height(4.dp))
        FileSizeText(spaceUsed, style = MaterialTheme.typography.headlineMedium)
        Text(
            pluralStringResource(R.plurals.home_file_count, totalFiles.toInt(), totalFiles),
            style = MaterialTheme.typography.bodySmall,
        )
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
private fun RecentFileItem(name: String, size: Long, mimeType: String, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Flat, radius = 12.dp)
            .clickable(onClick = onClick)
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
