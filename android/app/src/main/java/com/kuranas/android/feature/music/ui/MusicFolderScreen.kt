package com.kuranas.android.feature.music.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView

@Composable
fun MusicFolderScreen(
    folderKey: String,
    onNavigateBack: () -> Unit,
    onPlayTrack: () -> Unit,
) {
    val viewModel: MusicFolderViewModel = hiltViewModel()
    LaunchedEffect(folderKey) { viewModel.load(folderKey) }
    val tracks by viewModel.tracks

    val title = folderKey.trimEnd('/').substringAfterLast('/').ifBlank { folderKey }

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = title, leadingIcon = Icons.AutoMirrored.Filled.ArrowBack, onLeadingClick = onNavigateBack)
        when {
            tracks == null -> LoadingView()
            tracks.isNullOrEmpty() -> EmptyView(stringResource(R.string.music_folder_empty))
            else -> LazyColumn(contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                items(tracks ?: emptyList(), key = { it.id }) { track ->
                    TrackListItem(track = track, onClick = {
                        viewModel.play(track)
                        onPlayTrack()
                    })
                }
            }
        }
    }
}
