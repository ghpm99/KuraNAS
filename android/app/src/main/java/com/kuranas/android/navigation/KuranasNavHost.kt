package com.kuranas.android.navigation

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.kuranas.android.core.server.ServerState
import com.kuranas.android.core.ui.components.KNFrame
import com.kuranas.android.feature.connection.ui.ConnectionScreen
import com.kuranas.android.feature.diary.ui.DiaryScreen
import com.kuranas.android.feature.files.ui.FilesScreen
import com.kuranas.android.feature.home.ui.HomeScreen
import com.kuranas.android.feature.images.ui.ImageViewerScreen
import com.kuranas.android.feature.images.ui.ImagesScreen
import com.kuranas.android.feature.music.ui.MusicAlbumScreen
import com.kuranas.android.feature.music.ui.MusicArtistScreen
import com.kuranas.android.feature.music.ui.MusicPlayerScreen
import com.kuranas.android.feature.music.ui.MusicPlaylistScreen
import com.kuranas.android.feature.music.ui.MusicScreen
import com.kuranas.android.feature.notifications.ui.NotificationsScreen
import com.kuranas.android.feature.search.ui.SearchScreen
import com.kuranas.android.feature.settings.ui.SettingsScreen
import com.kuranas.android.feature.video.ui.VideoPlayerScreen
import com.kuranas.android.feature.video.ui.VideoPlaylistScreen
import com.kuranas.android.feature.video.ui.VideoScreen
import kotlinx.coroutines.launch

private const val ROUTE_HOME = "home"

@Composable
fun KuranasRoot(
    modifier: Modifier = Modifier,
    rootViewModel: RootViewModel = hiltViewModel(),
) {
    val serverState by rootViewModel.serverState.collectAsStateWithLifecycle()

    KNFrame(modifier = modifier) {
        when (serverState) {
            is ServerState.NotConfigured, is ServerState.Discovering -> ConnectionScreen()
            else -> AuthenticatedApp()
        }
    }
}

@Composable
private fun AuthenticatedApp() {
    val navController = rememberNavController()

    NavHost(navController = navController, startDestination = ROUTE_HOME, modifier = Modifier.fillMaxSize()) {
        composable(ROUTE_HOME) { AppPagerHost(navController) }

        composable(AppRoute.MUSIC_PLAYER) {
            MusicPlayerScreen(onNavigateBack = { navController.popBackStack() })
        }
        composable(
            route = AppRoute.VIDEO_PLAYER,
            arguments = listOf(navArgument("id") { type = NavType.StringType }),
        ) {
            VideoPlayerScreen(
                videoId = it.arguments?.getString("id") ?: "",
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
        composable(
            route = AppRoute.MUSIC_ARTIST,
            arguments = listOf(navArgument("key") { type = NavType.StringType }),
        ) {
            MusicArtistScreen(
                artistKey = it.arguments?.getString("key") ?: "",
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = { navController.navigate(AppRoute.MUSIC_PLAYER) },
            )
        }
        composable(
            route = AppRoute.MUSIC_ALBUM,
            arguments = listOf(navArgument("key") { type = NavType.StringType }),
        ) {
            MusicAlbumScreen(
                albumKey = it.arguments?.getString("key") ?: "",
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = { navController.navigate(AppRoute.MUSIC_PLAYER) },
            )
        }
        composable(
            route = AppRoute.MUSIC_PLAYLIST,
            arguments = listOf(navArgument("id") { type = NavType.IntType }),
        ) {
            MusicPlaylistScreen(
                playlistId = it.arguments?.getInt("id") ?: 0,
                onNavigateBack = { navController.popBackStack() },
                onPlayTrack = { navController.navigate(AppRoute.MUSIC_PLAYER) },
            )
        }
        composable(
            route = AppRoute.VIDEO_PLAYLIST,
            arguments = listOf(navArgument("id") { type = NavType.IntType }),
        ) {
            VideoPlaylistScreen(
                playlistId = it.arguments?.getInt("id") ?: 0,
                onNavigateBack = { navController.popBackStack() },
                onPlayVideo = { id -> navController.navigate(AppRoute.videoPlayer(id)) },
            )
        }
    }
}

@Composable
private fun AppPagerHost(navController: androidx.navigation.NavHostController) {
    val pages = SwipePage.entries
    val pagerState = rememberPagerState(pageCount = { pages.size })
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val serverStore = hiltViewModel<RootViewModel>()

    fun goToPage(page: SwipePage) {
        scope.launch { pagerState.animateScrollToPage(page.ordinal) }
    }

    ModalNavigationDrawer(
        drawerState = drawerState,
        gesturesEnabled = pagerState.currentPage == 0 || drawerState.isOpen,
        drawerContent = {
            KNDrawer(
                current = pagerState.currentPage,
                onSelect = { index ->
                    scope.launch {
                        drawerState.close()
                        pagerState.scrollToPage(index)
                    }
                },
                onForget = { /* handled in SettingsScreen */ },
            )
        },
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize(),
            beyondViewportPageCount = 1,
        ) { index ->
            when (pages[index]) {
                SwipePage.HOME -> HomeScreen(
                    onOpenMenu = { scope.launch { drawerState.open() } },
                    onOpenFiles = { goToPage(SwipePage.FILES) },
                    onOpenMusic = { goToPage(SwipePage.MUSIC) },
                    onOpenVideo = { goToPage(SwipePage.VIDEO) },
                    onOpenImages = { goToPage(SwipePage.IMAGES) },
                )
                SwipePage.FILES -> FilesScreen(
                    onOpenImage = { id -> navController.navigate(AppRoute.imageViewer(id)) },
                    onOpenVideo = { id -> navController.navigate(AppRoute.videoPlayer(id)) },
                    onPlayAudio = { navController.navigate(AppRoute.MUSIC_PLAYER) },
                )
                SwipePage.MUSIC -> MusicScreen(
                    onOpenPlayer = { navController.navigate(AppRoute.MUSIC_PLAYER) },
                    onOpenArtist = { key -> navController.navigate(AppRoute.musicArtist(key)) },
                    onOpenAlbum = { key -> navController.navigate(AppRoute.musicAlbum(key)) },
                    onOpenPlaylist = { id -> navController.navigate(AppRoute.musicPlaylist(id)) },
                )
                SwipePage.VIDEO -> VideoScreen(
                    onPlayVideo = { id -> navController.navigate(AppRoute.videoPlayer(id)) },
                    onOpenPlaylist = { id -> navController.navigate(AppRoute.videoPlaylist(id)) },
                )
                SwipePage.IMAGES -> ImagesScreen(
                    onOpenImage = { id -> navController.navigate(AppRoute.imageViewer(id)) },
                )
                SwipePage.SEARCH -> SearchScreen(
                    onOpenFile = { id -> navController.navigate(AppRoute.fileDetail(id)) },
                    onPlayAudio = { navController.navigate(AppRoute.MUSIC_PLAYER) },
                    onPlayVideo = { id -> navController.navigate(AppRoute.videoPlayer(id)) },
                )
                SwipePage.DIARY -> DiaryScreen()
                SwipePage.NOTIFICATIONS -> NotificationsScreen()
                SwipePage.SETTINGS -> SettingsScreen()
            }
        }
    }
}
