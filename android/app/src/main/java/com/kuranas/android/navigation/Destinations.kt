package com.kuranas.android.navigation

import androidx.annotation.StringRes
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Book
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.MusicNote
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.VideoLibrary
import androidx.compose.ui.graphics.vector.ImageVector
import com.kuranas.android.R

enum class SwipePage(@StringRes val labelRes: Int, val icon: ImageVector) {
    HOME(R.string.nav_home, Icons.Filled.Home),
    FILES(R.string.nav_files, Icons.Filled.Folder),
    MUSIC(R.string.nav_music, Icons.Filled.MusicNote),
    VIDEO(R.string.nav_videos, Icons.Filled.VideoLibrary),
    IMAGES(R.string.nav_images, Icons.Filled.PhotoLibrary),
    SEARCH(R.string.nav_search, Icons.Filled.Search),
    DIARY(R.string.nav_diary, Icons.Filled.Book),
    NOTIFICATIONS(R.string.nav_notifications, Icons.Filled.Notifications),
    SETTINGS(R.string.nav_settings, Icons.Filled.Settings),
}

object AppRoute {
    const val FILE_DETAIL = "file/{id}?name={name}"
    // Navegação sem argumento (mini-player, telas de música) cai no default trackId=-1.
    const val MUSIC_PLAYER = "music-player"
    const val MUSIC_PLAYER_ROUTE = "music-player?trackId={trackId}"
    const val VIDEO_PLAYER = "video-player/{id}"
    const val IMAGE_VIEWER = "image-viewer/{id}"
    const val MUSIC_ARTIST = "music/artist/{key}"
    const val MUSIC_ALBUM = "music/album/{key}"
    const val MUSIC_FOLDER = "music/folder/{key}"
    const val MUSIC_PLAYLIST = "music/playlist/{id}"
    const val VIDEO_PLAYLIST = "video/playlist/{id}"
    const val JOBS = "jobs"
    const val JOB_DETAIL = "jobs/{id}"
    const val ANALYTICS = "analytics"

    fun fileDetail(id: String, name: String = "") =
        "file/$id?name=${android.net.Uri.encode(name)}"
    fun musicPlayer(trackId: Int) = "music-player?trackId=$trackId"
    fun videoPlayer(id: String) = "video-player/$id"
    fun imageViewer(id: String) = "image-viewer/$id"
    fun musicArtist(key: String) = "music/artist/$key"
    fun musicAlbum(key: String) = "music/album/$key"
    // A pasta é um caminho com barras; codifica para virar um único argumento de rota.
    fun musicFolder(key: String) = "music/folder/${android.net.Uri.encode(key)}"
    fun musicPlaylist(id: Int) = "music/playlist/$id"
    fun videoPlaylist(id: Int) = "video/playlist/$id"
    fun jobDetail(id: String) = "jobs/$id"
}
