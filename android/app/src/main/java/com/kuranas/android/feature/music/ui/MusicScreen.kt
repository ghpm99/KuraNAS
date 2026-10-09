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
import androidx.compose.foundation.lazy.rememberLazyListState
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
import androidx.compose.ui.res.pluralStringResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
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
    val tab by viewModel.tab.collectAsStateWithLifecycle()
    val tracks by viewModel.tracks.collectAsStateWithLifecycle()
    val artists by viewModel.artists.collectAsStateWithLifecycle()
    val albums by viewModel.albums.collectAsStateWithLifecycle()
    val playlists by viewModel.playlists.collectAsStateWithLifecycle()
    val folders by viewModel.folders.collectAsStateWithLifecycle()
    val currentList = when (tab) {
        MusicTab.TRACKS -> tracks.toListStatus()
        MusicTab.ARTISTS -> artists.toListStatus()
        MusicTab.ALBUMS -> albums.toListStatus()
        MusicTab.PLAYLISTS -> playlists.toListStatus()
        MusicTab.FOLDERS -> folders.toListStatus()
    }
    val tabs = MusicTab.entries

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = stringResource(R.string.nav_music))
        ScrollableTabRow(
            selectedTabIndex = tab.ordinal,
            containerColor = androidx.compose.ui.graphics.Color.Transparent,
        ) {
            tabs.forEach { tabEntry ->
                Tab(
                    selected = tab == tabEntry,
                    onClick = { viewModel.selectTab(tabEntry) },
                    text = {
                        Text(
                            when (tabEntry) {
                                MusicTab.TRACKS -> stringResource(R.string.music_tab_tracks)
                                MusicTab.ARTISTS -> stringResource(R.string.music_tab_artists)
                                MusicTab.ALBUMS -> stringResource(R.string.music_tab_albums)
                                MusicTab.PLAYLISTS -> stringResource(R.string.music_tab_playlists)
                                MusicTab.FOLDERS -> stringResource(R.string.music_tab_folders)
                            }
                        )
                    },
                )
            }
        }

        PullToRefreshBox(
            isRefreshing = currentList.isRefreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier.fillMaxSize(),
        ) {
            when {
                currentList.isLoading -> LoadingView()
                currentList.error != null -> ErrorView(currentList.error)
                else -> when (tab) {
                    MusicTab.ARTISTS -> ArtistsList(artists, onOpenArtist, viewModel::loadMoreArtists)
                    MusicTab.ALBUMS -> AlbumsList(albums, onOpenAlbum, viewModel::loadMoreAlbums)
                    MusicTab.PLAYLISTS -> PlaylistsList(playlists, onOpenPlaylist, viewModel::loadMorePlaylists)
                    MusicTab.FOLDERS -> FoldersList(folders, onOpenFolder, viewModel::loadMoreFolders)
                    MusicTab.TRACKS -> TracksList(tracks, viewModel::loadMoreTracks) { track ->
                        viewModel.play(track, tracks.items)
                        onOpenPlayer()
                    }
                }
            }
        }
    }
}

private data class ListStatus(val isLoading: Boolean, val isRefreshing: Boolean, val error: String?)

private fun PagedListState<*>.toListStatus() = ListStatus(isLoading, isRefreshing, error)

@Composable
private fun ArtistsList(pagedState: PagedListState<ArtistDto>, onOpen: (String) -> Unit, onLoadMore: () -> Unit) {
    val artists = pagedState.items
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, artists.size, pagedState.hasMore, onLoadMore)
    if (artists.isEmpty()) { EmptyView(stringResource(R.string.music_no_artists)); return }
    LazyColumn(state = listState, contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
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
                        Text(pluralStringResource(R.plurals.music_track_count, artist.trackCount, artist.trackCount), style = MaterialTheme.typography.bodySmall)
                    }
                }
                Icon(Icons.Default.ChevronRight, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}

@Composable
private fun AlbumsList(pagedState: PagedListState<AlbumDto>, onOpen: (String) -> Unit, onLoadMore: () -> Unit) {
    val albums = pagedState.items
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, albums.size, pagedState.hasMore, onLoadMore)
    if (albums.isEmpty()) { EmptyView(stringResource(R.string.music_no_albums)); return }
    LazyColumn(state = listState, contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
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
                        Text(album.artist.ifBlank { stringResource(R.string.music_unknown_artist) }, style = MaterialTheme.typography.bodySmall)
                    }
                }
                Text("${album.trackCount}", style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun TracksList(pagedState: PagedListState<TrackDto>, onLoadMore: () -> Unit, onPlay: (TrackDto) -> Unit) {
    val tracks = pagedState.items
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, tracks.size, pagedState.hasMore, onLoadMore)
    if (tracks.isEmpty()) { EmptyView(stringResource(R.string.music_no_tracks)); return }
    LazyColumn(state = listState, contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items(tracks, key = { it.id }) { track ->
            TrackListItem(track = track, onClick = { onPlay(track) })
        }
    }
}

@Composable
private fun FoldersList(pagedState: PagedListState<FolderDto>, onOpen: (String) -> Unit, onLoadMore: () -> Unit) {
    val folders = pagedState.items
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, folders.size, pagedState.hasMore, onLoadMore)
    if (folders.isEmpty()) { EmptyView(stringResource(R.string.music_no_folders)); return }
    LazyColumn(state = listState, contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
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
                Text(pluralStringResource(R.plurals.music_track_count, folder.trackCount, folder.trackCount), style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun PlaylistsList(pagedState: PagedListState<PlaylistDto>, onOpen: (Int) -> Unit, onLoadMore: () -> Unit) {
    val playlists = pagedState.items
    val listState = rememberLazyListState()
    LoadMoreOnNearEnd(listState, playlists.size, pagedState.hasMore, onLoadMore)
    if (playlists.isEmpty()) { EmptyView(stringResource(R.string.music_no_playlists)); return }
    LazyColumn(state = listState, contentPadding = PaddingValues(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
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
                Text(pluralStringResource(R.plurals.music_track_count, playlist.trackCount, playlist.trackCount), style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}
