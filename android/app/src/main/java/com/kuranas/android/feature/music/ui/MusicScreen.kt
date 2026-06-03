package com.kuranas.android.feature.music.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Album
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.PlaylistPlay
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ScrollableTabRow
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
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
import com.kuranas.android.feature.music.data.AlbumDto
import com.kuranas.android.feature.music.data.ArtistDto
import com.kuranas.android.feature.music.data.FolderDto
import com.kuranas.android.feature.music.data.PlaylistDto
import com.kuranas.android.feature.music.data.TrackDto

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MusicScreen(
    onOpenPlayer: () -> Unit,
    onOpenArtist: (String) -> Unit,
    onOpenAlbum: (String) -> Unit,
    onOpenPlaylist: (Int) -> Unit,
    onOpenFolder: (String) -> Unit,
    viewModel: MusicViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val tabs = MusicTab.entries

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = "Música")
        ScrollableTabRow(
            selectedTabIndex = state.tab.ordinal,
            containerColor = androidx.compose.ui.graphics.Color.Transparent,
        ) {
            tabs.forEach { tab ->
                Tab(
                    selected = state.tab == tab,
                    onClick = { viewModel.selectTab(tab) },
                    text = {
                        Text(
                            when (tab) {
                                MusicTab.TRACKS -> "Faixas"
                                MusicTab.ARTISTS -> "Artistas"
                                MusicTab.ALBUMS -> "Álbuns"
                                MusicTab.PLAYLISTS -> "Playlists"
                                MusicTab.FOLDERS -> "Pastas"
                            }
                        )
                    },
                )
            }
        }

        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
            when {
                state.isLoading -> LoadingView()
                state.error != null -> ErrorView(state.error!!)
                else -> when (state.tab) {
                    MusicTab.ARTISTS -> ArtistsList(state.artists, onOpenArtist)
                    MusicTab.ALBUMS -> AlbumsList(state.albums, onOpenAlbum)
                    MusicTab.PLAYLISTS -> PlaylistsList(state.playlists, onOpenPlaylist)
                    MusicTab.FOLDERS -> FoldersList(state.folders, onOpenFolder)
                    MusicTab.TRACKS -> TracksList(state.tracks) { track ->
                        viewModel.play(track, state.tracks)
                        onOpenPlayer()
                    }
                }
            }
        }
    }
}

@Composable
private fun ArtistsList(artists: List<ArtistDto>, onOpen: (String) -> Unit) {
    if (artists.isEmpty()) { EmptyView("Nenhum artista encontrado"); return }
    LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(artists, key = { it.key }) { artist ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .glass(GlassLevel.Flat, radius = 12.dp)
                    .clickable { onOpen(artist.key) }
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Icon(Icons.Default.Person, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
                    Column {
                        Text(artist.artist, style = MaterialTheme.typography.bodyMedium)
                        Text("${artist.trackCount} faixas", style = MaterialTheme.typography.bodySmall)
                    }
                }
                Icon(Icons.Default.ChevronRight, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}

@Composable
private fun AlbumsList(albums: List<AlbumDto>, onOpen: (String) -> Unit) {
    if (albums.isEmpty()) { EmptyView("Nenhum álbum encontrado"); return }
    LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(albums, key = { it.key }) { album ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .glass(GlassLevel.Flat, radius = 12.dp)
                    .clickable { onOpen(album.key) }
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Icon(Icons.Default.Album, contentDescription = null, tint = MaterialTheme.colorScheme.secondary)
                    Column {
                        Text(album.album, style = MaterialTheme.typography.bodyMedium)
                        Text(album.artist.ifBlank { "Artista desconhecido" }, style = MaterialTheme.typography.bodySmall)
                    }
                }
                Text("${album.trackCount}", style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun TracksList(tracks: List<TrackDto>, onPlay: (TrackDto) -> Unit) {
    if (tracks.isEmpty()) { EmptyView("Nenhuma faixa"); return }
    LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(tracks, key = { it.id }) { track ->
            TrackListItem(track = track, onClick = { onPlay(track) })
        }
    }
}

@Composable
private fun FoldersList(folders: List<FolderDto>, onOpen: (String) -> Unit) {
    if (folders.isEmpty()) { EmptyView("Nenhuma pasta com música"); return }
    LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(folders, key = { it.folder }) { folder ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .glass(GlassLevel.Flat, radius = 12.dp)
                    .clickable { onOpen(folder.folder) }
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Icon(Icons.Default.Folder, contentDescription = null, tint = MaterialTheme.colorScheme.secondary)
                    Text(
                        folder.folder.trimEnd('/').substringAfterLast('/').ifBlank { folder.folder },
                        style = MaterialTheme.typography.bodyMedium,
                    )
                }
                Text("${folder.trackCount} faixas", style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun PlaylistsList(playlists: List<PlaylistDto>, onOpen: (Int) -> Unit) {
    if (playlists.isEmpty()) { EmptyView("Nenhuma playlist"); return }
    LazyColumn(contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(playlists, key = { it.id }) { playlist ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .glass(GlassLevel.Flat, radius = 12.dp)
                    .clickable { onOpen(playlist.id) }
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Icon(Icons.Default.PlaylistPlay, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
                    Text(playlist.name, style = MaterialTheme.typography.bodyMedium)
                }
                Text("${playlist.trackCount} faixas", style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}
