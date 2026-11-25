package com.kuranas.android.navigation

import androidx.annotation.StringRes
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.MoreHoriz
import androidx.compose.material.icons.filled.MusicNote
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material.icons.filled.VideoLibrary
import androidx.compose.ui.graphics.vector.ImageVector
import com.kuranas.android.R

enum class BottomTab(
    val graphRoute: String,
    val startRoute: String,
    @StringRes val labelRes: Int,
    val icon: ImageVector,
) {
    HOME("tab-home", AppRoute.HOME, R.string.nav_home, Icons.Filled.Home),
    FILES("tab-files", AppRoute.FILES, R.string.nav_files, Icons.Filled.Folder),
    IMAGES("tab-images", AppRoute.IMAGES, R.string.nav_images, Icons.Filled.PhotoLibrary),
    MUSIC("tab-music", AppRoute.MUSIC, R.string.nav_music, Icons.Filled.MusicNote),
    VIDEO("tab-video", AppRoute.VIDEO, R.string.nav_videos, Icons.Filled.VideoLibrary),
    MORE("tab-more", AppRoute.MORE, R.string.nav_more, Icons.Filled.MoreHoriz),
}

object AppRoute {
    const val FILE_DETAIL = "file/{id}?name={name}"
    const val MUSIC_PLAYER = "music-player"
    const val MUSIC_PLAYER_ROUTE = "music-player?trackId={trackId}"
    const val VIDEO_PLAYER = "video-player/{id}?playlistId={playlistId}"
    const val NO_PLAYLIST_ARGUMENT = -1
    const val IMAGE_VIEWER = "image-viewer/{id}"
    const val MUSIC_ARTIST = "music/artist/{key}"
    const val MUSIC_ALBUM = "music/album/{key}"
    const val MUSIC_FOLDER = "music/folder/{key}"
    const val MUSIC_PLAYLIST = "music/playlist/{id}"
    const val VIDEO_PLAYLIST = "video/playlist/{id}"
    const val VIDEO_LIBRARY = "video/library"
    const val FILES_FOLDER = "files/folder/{folderId}?name={name}"
    const val TRANSFERS = "transfers"
    const val CAMERA_UPLOAD = "camera-upload"
    const val JOBS = "jobs"
    const val INGEST = "ingest"
    const val JOB_DETAIL = "jobs/{id}"
    const val HOME = "home"
    const val FILES = "files"
    const val IMAGES = "images"
    const val MUSIC = "music"
    const val VIDEO = "video"
    const val MORE = "more"
    const val DOCUMENTS = "documents"
    const val FAVORITES = "favorites"
    const val SEARCH = "search"
    const val TRASH = "trash"
    const val NOTIFICATIONS = "notifications"
    const val DIARY = "diary"
    const val SETTINGS = "settings"

    fun fileDetail(id: String, name: String = "") =
        "file/$id?name=${android.net.Uri.encode(name)}"
    fun filesFolder(id: String, name: String) = "files/folder/$id?name=${android.net.Uri.encode(name)}"
    fun musicPlayer(trackId: Int) = "music-player?trackId=$trackId"
    fun videoPlayer(id: String, playlistId: Int? = null) =
        if (playlistId == null) "video-player/$id" else "video-player/$id?playlistId=$playlistId"
    fun imageViewer(id: String) = "image-viewer/$id"
    fun musicArtist(key: String) = "music/artist/$key"
    fun musicAlbum(key: String) = "music/album/$key"
    fun musicFolder(key: String) = "music/folder/${android.net.Uri.encode(key)}"
    fun musicPlaylist(id: Int) = "music/playlist/$id"
    fun videoPlaylist(id: Int) = "video/playlist/$id"
    fun jobDetail(id: String) = "jobs/$id"
}
