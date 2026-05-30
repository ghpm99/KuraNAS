package com.kuranas.android.feature.search.ui

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
import androidx.compose.material.icons.filled.AudioFile
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Image
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.VideoFile
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.search.data.SearchFileDto

@Composable
fun SearchScreen(
    onOpenFile: (String) -> Unit,
    onPlayAudio: (String) -> Unit,
    onPlayVideo: (String) -> Unit,
    viewModel: SearchViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = "Busca")
        TextField(
            value = state.query,
            onValueChange = viewModel::onQueryChange,
            placeholder = { Text("Buscar arquivos, músicas, vídeos...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            singleLine = true,
            colors = TextFieldDefaults.colors(
                focusedContainerColor = Color.Transparent,
                unfocusedContainerColor = Color.Transparent,
            ),
            modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light, radius = 16.dp),
        )

        when {
            state.isLoading -> LoadingView()
            state.results != null -> {
                val results = state.results!!
                LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    if (results.files.isNotEmpty()) {
                        item { Text("Arquivos", style = MaterialTheme.typography.labelMedium, modifier = Modifier.padding(top = 8.dp)) }
                        items(results.files) { file ->
                            SearchResultItem(file = file, onClick = {
                                when {
                                    file.mimeType.startsWith("audio/") -> onPlayAudio(file.id)
                                    file.mimeType.startsWith("video/") -> onPlayVideo(file.id)
                                    else -> onOpenFile(file.id)
                                }
                            })
                        }
                    }
                    if (results.music.isNotEmpty()) {
                        item { Text("Música", style = MaterialTheme.typography.labelMedium, modifier = Modifier.padding(top = 8.dp)) }
                        items(results.music) { file ->
                            SearchResultItem(file = file, onClick = { onPlayAudio(file.id) })
                        }
                    }
                    if (results.videos.isNotEmpty()) {
                        item { Text("Vídeos", style = MaterialTheme.typography.labelMedium, modifier = Modifier.padding(top = 8.dp)) }
                        items(results.videos) { file ->
                            SearchResultItem(file = file, onClick = { onPlayVideo(file.id) })
                        }
                    }
                    if (results.total == 0) {
                        item { Text("Nenhum resultado para \"${state.query}\"", style = MaterialTheme.typography.bodyMedium, modifier = Modifier.padding(top = 16.dp)) }
                    }
                }
            }
        }
    }
}

@Composable
private fun SearchResultItem(file: SearchFileDto, onClick: () -> Unit) {
    val icon = when {
        file.isDir -> Icons.Default.Folder
        file.mimeType.startsWith("image/") -> Icons.Default.Image
        file.mimeType.startsWith("video/") -> Icons.Default.VideoFile
        file.mimeType.startsWith("audio/") -> Icons.Default.AudioFile
        else -> Icons.Default.InsertDriveFile
    }
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Flat, radius = 12.dp)
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(icon, contentDescription = null, modifier = Modifier.size(20.dp), tint = MaterialTheme.colorScheme.primary)
        Column {
            Text(file.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1)
            Text(file.path, style = MaterialTheme.typography.bodySmall, maxLines = 1)
        }
    }
}
