package com.kuranas.android.navigation

import androidx.activity.ComponentActivity
import androidx.activity.compose.LocalActivity
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavHostController
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.compose.navigation
import androidx.navigation.navArgument
import com.kuranas.android.feature.cameraupload.ui.CameraUploadScreen
import com.kuranas.android.feature.diary.ui.DiaryScreen
import com.kuranas.android.feature.files.ui.FOLDER_ID_ARGUMENT
import com.kuranas.android.feature.files.ui.FOLDER_NAME_ARGUMENT
import com.kuranas.android.feature.files.ui.FavoritesScreen
import com.kuranas.android.feature.files.ui.FileViewerScreen
import com.kuranas.android.feature.files.ui.FilesScreen
import com.kuranas.android.feature.home.ui.HomeScreen
import com.kuranas.android.feature.images.ui.ImageViewerScreen
import com.kuranas.android.feature.images.ui.ImagesScreen
import com.kuranas.android.feature.jobs.ui.JobsScreen
import com.kuranas.android.feature.music.ui.MusicAlbumScreen
import com.kuranas.android.feature.music.ui.MusicArtistScreen
import com.kuranas.android.feature.music.ui.MusicFolderScreen
import com.kuranas.android.feature.music.ui.MusicPlayerScreen
import com.kuranas.android.feature.music.ui.MusicPlaylistScreen
import com.kuranas.android.feature.music.ui.MusicScreen
import com.kuranas.android.feature.notifications.ui.NotificationsIndicatorViewModel
import com.kuranas.android.feature.notifications.ui.NotificationsScreen
import com.kuranas.android.feature.search.ui.SearchScreen
import com.kuranas.android.feature.settings.ui.SettingsScreen
import com.kuranas.android.feature.trash.ui.TrashScreen
import com.kuranas.android.feature.transfers.ui.TransfersIndicatorViewModel
import com.kuranas.android.feature.transfers.ui.TransfersScreen
import com.kuranas.android.feature.video.ui.VideoLibraryScreen
import com.kuranas.android.feature.video.ui.VideoPlayerScreen
import com.kuranas.android.feature.video.ui.VideoPlaylistScreen
import com.kuranas.android.feature.video.ui.VideoScreen

fun NavHostController.openTab(tab: BottomTab) {
    navigate(tab.graphRoute) {
        popUpTo(graph.findStartDestination().id) { saveState = true }
        launchSingleTop = true
        restoreState = true
    }
}

fun NavHostController.openTransfers() {
    navigate(AppRoute.TRANSFERS) { launchSingleTop = true }
}

private fun NavHostController.openSearch() = navigate(AppRoute.SEARCH) { launchSingleTop = true }

private fun NavHostController.openImage(id: String) = navigate(AppRoute.imageViewer(id))

private fun NavHostController.openVideo(id: String) = navigate(AppRoute.videoPlayer(id))

private fun NavHostController.playAudio(id: String) = navigate(AppRoute.musicPlayer(id.toIntOrNull() ?: -1))

private fun NavHostController.openFile(id: String, name: String) = navigate(AppRoute.fileDetail(id, name))

private fun NavHostController.openMusicPlayer() = navigate(AppRoute.MUSIC_PLAYER)

fun NavGraphBuilder.tabGraphs(navController: NavHostController) {
    homeGraph(navController)
    filesGraph(navController)
    imagesGraph(navController)
    musicGraph(navController)
    videoGraph(navController)
    moreGraph(navController)
}

private fun NavGraphBuilder.homeGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.HOME.startRoute, route = BottomTab.HOME.graphRoute) {
        composable(AppRoute.HOME) {
            HomeScreen(
                onOpenFiles = { navController.openTab(BottomTab.FILES) },
                onOpenMusic = { navController.openTab(BottomTab.MUSIC) },
                onOpenVideo = { navController.openTab(BottomTab.VIDEO) },
                onOpenImages = { navController.openTab(BottomTab.IMAGES) },
                onOpenImage = navController::openImage,
                onOpenVideoFile = navController::openVideo,
                onPlayAudio = navController::playAudio,
                onOpenFile = navController::openFile,
                onOpenSearch = navController::openSearch,
            )
        }
    }
}

private fun NavGraphBuilder.filesGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.FILES.startRoute, route = BottomTab.FILES.graphRoute) {
        composable(AppRoute.FILES) { FilesDestination(navController) }
    }
}

private fun NavGraphBuilder.imagesGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.IMAGES.startRoute, route = BottomTab.IMAGES.graphRoute) {
        composable(AppRoute.IMAGES) { ImagesScreen(onOpenImage = navController::openImage, onOpenSearch = navController::openSearch) }
    }
}

private fun NavGraphBuilder.musicGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.MUSIC.startRoute, route = BottomTab.MUSIC.graphRoute) {
        composable(AppRoute.MUSIC) {
            MusicScreen(
                onOpenPlayer = navController::openMusicPlayer,
                onOpenArtist = { key -> navController.navigate(AppRoute.musicArtist(key)) },
                onOpenAlbum = { key -> navController.navigate(AppRoute.musicAlbum(key)) },
                onOpenPlaylist = { id -> navController.navigate(AppRoute.musicPlaylist(id)) },
                onOpenFolder = { key -> navController.navigate(AppRoute.musicFolder(key)) },
                onOpenSearch = navController::openSearch,
            )
        }
        composable(
            route = AppRoute.MUSIC_ARTIST,
            arguments = listOf(navArgument("key") { type = NavType.StringType }),
        ) {
            MusicArtistScreen(
                artistKey = it.arguments?.getString("key") ?: "",
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = navController::openMusicPlayer,
            )
        }
        composable(
            route = AppRoute.MUSIC_ALBUM,
            arguments = listOf(navArgument("key") { type = NavType.StringType }),
        ) {
            MusicAlbumScreen(
                albumKey = it.arguments?.getString("key") ?: "",
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = navController::openMusicPlayer,
            )
        }
        composable(
            route = AppRoute.MUSIC_FOLDER,
            arguments = listOf(navArgument("key") { type = NavType.StringType }),
        ) {
            MusicFolderScreen(
                folderKey = it.arguments?.getString("key") ?: "",
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = navController::openMusicPlayer,
            )
        }
        composable(
            route = AppRoute.MUSIC_PLAYLIST,
            arguments = listOf(navArgument("id") { type = NavType.IntType }),
        ) {
            MusicPlaylistScreen(
                playlistId = it.arguments?.getInt("id") ?: 0,
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = navController::openMusicPlayer,
            )
        }
    }
}

private fun NavGraphBuilder.videoGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.VIDEO.startRoute, route = BottomTab.VIDEO.graphRoute) {
        composable(AppRoute.VIDEO) {
            VideoScreen(
                onPlayVideo = navController::openVideo,
                onOpenPlaylist = { id -> navController.navigate(AppRoute.videoPlaylist(id)) },
                onOpenLibrary = { navController.navigate(AppRoute.VIDEO_LIBRARY) },
                onOpenSearch = navController::openSearch,
            )
        }
        composable(
            route = AppRoute.VIDEO_PLAYLIST,
            arguments = listOf(navArgument("id") { type = NavType.IntType }),
        ) {
            VideoPlaylistScreen(
                playlistId = it.arguments?.getInt("id") ?: 0,
                onNavigateBack = { navController.popBackStack() },
                onPlayVideo = { id -> navController.navigate(AppRoute.videoPlayer(id, it.arguments?.getInt("id"))) },
            )
        }
        composable(AppRoute.VIDEO_LIBRARY) {
            VideoLibraryScreen(
                onNavigateBack = { navController.popBackStack() },
                onPlayVideo = navController::openVideo,
            )
        }
    }
}

private fun NavGraphBuilder.moreGraph(navController: NavHostController) {
    navigation(startDestination = BottomTab.MORE.startRoute, route = BottomTab.MORE.graphRoute) {
        composable(AppRoute.MORE) { MoreDestination(navController) }
        composable(AppRoute.FAVORITES) {
            FavoritesScreen(
                onNavigateBack = { navController.popBackStack() },
                onOpenFolder = { id, name -> navController.navigate(AppRoute.filesFolder(id, name)) },
                onOpenImage = navController::openImage,
                onOpenVideo = navController::openVideo,
                onPlayAudio = navController::playAudio,
                onOpenFile = navController::openFile,
            )
        }
        composable(AppRoute.TRASH) { TrashScreen(onNavigateBack = { navController.popBackStack() }) }
        composable(AppRoute.NOTIFICATIONS) { NotificationsScreen(onNavigateBack = { navController.popBackStack() }) }
        composable(AppRoute.JOBS) { JobsScreen(onNavigateBack = { navController.popBackStack() }) }
        composable(AppRoute.DIARY) { DiaryScreen(onNavigateBack = { navController.popBackStack() }) }
        composable(AppRoute.SETTINGS) {
            SettingsScreen(
                onNavigateBack = { navController.popBackStack() },
                onOpenCameraUpload = { navController.navigate(AppRoute.CAMERA_UPLOAD) { launchSingleTop = true } },
            )
        }
        composable(
            route = AppRoute.FILES_FOLDER,
            arguments = listOf(
                navArgument(FOLDER_ID_ARGUMENT) { type = NavType.StringType },
                navArgument(FOLDER_NAME_ARGUMENT) { type = NavType.StringType; defaultValue = "" },
            ),
        ) { FilesDestination(navController) }
    }
}

fun NavGraphBuilder.topLevelRoutes(navController: NavHostController) {
    composable(
        route = AppRoute.MUSIC_PLAYER_ROUTE,
        arguments = listOf(navArgument("trackId") { type = NavType.IntType; defaultValue = -1 }),
    ) {
        MusicPlayerScreen(onNavigateBack = { navController.popBackStack() })
    }
    composable(
        route = AppRoute.FILE_DETAIL,
        arguments = listOf(
            navArgument("id") { type = NavType.StringType },
            navArgument("name") { type = NavType.StringType; defaultValue = "" },
        ),
    ) {
        FileViewerScreen(
            fileId = it.arguments?.getString("id") ?: "",
            fileName = it.arguments?.getString("name") ?: "",
            onNavigateBack = { navController.popBackStack() },
        )
    }
    composable(
        route = AppRoute.VIDEO_PLAYER,
        arguments = listOf(
            navArgument("id") { type = NavType.StringType },
            navArgument("playlistId") { type = NavType.IntType; defaultValue = AppRoute.NO_PLAYLIST_ARGUMENT },
        ),
    ) {
        VideoPlayerScreen(
            videoId = it.arguments?.getString("id") ?: "",
            playlistId = it.arguments?.getInt("playlistId")?.takeIf { id -> id != AppRoute.NO_PLAYLIST_ARGUMENT },
            onNavigateBack = { navController.popBackStack() },
        )
    }
    composable(
        route = AppRoute.IMAGE_VIEWER,
        arguments = listOf(navArgument("id") { type = NavType.StringType }),
    ) {
        ImageViewerScreen(
            fileId = it.arguments?.getString("id") ?: "",
            onNavigateBack = { navController.popBackStack() },
        )
    }
    composable(AppRoute.SEARCH) {
        SearchScreen(
            onNavigateBack = { navController.popBackStack() },
            onOpenFile = { id -> navController.navigate(AppRoute.fileDetail(id)) },
            onPlayAudio = navController::playAudio,
            onPlayVideo = navController::openVideo,
        )
    }
    composable(AppRoute.TRANSFERS) {
        TransfersScreen(onNavigateBack = { navController.popBackStack() })
    }
    composable(AppRoute.CAMERA_UPLOAD) {
        CameraUploadScreen(
            onNavigateBack = { navController.popBackStack() },
            onOpenTransfers = navController::openTransfers,
        )
    }
}

@Composable
private fun FilesDestination(navController: NavHostController) {
    FilesScreen(
        onOpenImage = navController::openImage,
        onOpenVideo = navController::openVideo,
        onPlayAudio = navController::playAudio,
        onOpenFile = navController::openFile,
        onOpenTransfers = navController::openTransfers,
        onOpenSearch = navController::openSearch,
    )
}

@Composable
private fun MoreDestination(navController: NavHostController) {
    val activityOwner = LocalActivity.current as ComponentActivity
    val transfersIndicatorViewModel: TransfersIndicatorViewModel = hiltViewModel(activityOwner)
    val notificationsIndicatorViewModel: NotificationsIndicatorViewModel = hiltViewModel(activityOwner)
    val activeTransfersCount by transfersIndicatorViewModel.activeCount.collectAsStateWithLifecycle()
    val unreadNotificationsCount by notificationsIndicatorViewModel.unreadCount.collectAsStateWithLifecycle()
    MoreScreen(
        activeTransfersCount = activeTransfersCount,
        unreadNotificationsCount = unreadNotificationsCount,
        onOpenEntry = { entry -> navController.navigate(entry.route) { launchSingleTop = true } },
    )
}
