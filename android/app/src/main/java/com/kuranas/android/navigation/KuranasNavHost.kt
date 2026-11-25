package com.kuranas.android.navigation

import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.LocalActivity
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Badge
import androidx.compose.material3.BadgedBox
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavDestination
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.kuranas.android.core.server.ServerConnectionState
import com.kuranas.android.core.ui.LocalServerReconnections
import com.kuranas.android.core.ui.components.KNFrame
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.feature.connection.ui.ConnectionScreen
import com.kuranas.android.feature.music.ui.MiniPlayer
import com.kuranas.android.feature.notifications.ui.NotificationsIndicatorViewModel
import com.kuranas.android.feature.transfers.ui.TransfersIndicatorViewModel
import com.kuranas.android.ui.theme.BgCanvas

@Composable
fun KuranasRoot(
    modifier: Modifier = Modifier,
    shouldOpenTransfers: Boolean = false,
    onOpenTransfersHandled: () -> Unit = {},
    requestedIngestUrl: String? = null,
    onIngestUrlHandled: () -> Unit = {},
    rootViewModel: RootViewModel = hiltViewModel(),
) {
    val destination by rootViewModel.destination.collectAsStateWithLifecycle()

    KNFrame(modifier = modifier) {
        when (val currentDestination = destination) {
            RootDestination.Checking -> LoadingView()
            RootDestination.Connection -> ConnectionScreen()
            RootDestination.App -> AuthenticatedApp(shouldOpenTransfers, onOpenTransfersHandled, requestedIngestUrl, onIngestUrlHandled)
            is RootDestination.AccessDenied -> AccessDeniedScreen(currentDestination.denial)
        }
    }
}

@Composable
private fun AuthenticatedApp(
    shouldOpenTransfers: Boolean,
    onOpenTransfersHandled: () -> Unit,
    requestedIngestUrl: String?,
    onIngestUrlHandled: () -> Unit,
) {
    val navController = rememberNavController()
    val activityOwner = LocalActivity.current as ComponentActivity
    val transfersIndicatorViewModel: TransfersIndicatorViewModel = hiltViewModel(activityOwner)
    val notificationsIndicatorViewModel: NotificationsIndicatorViewModel = hiltViewModel(activityOwner)
    val connectionBannerViewModel: ConnectionBannerViewModel = hiltViewModel()
    val connectionState by connectionBannerViewModel.connectionState.collectAsStateWithLifecycle()
    val activeTransfersCount by transfersIndicatorViewModel.activeCount.collectAsStateWithLifecycle()
    val unreadNotificationsCount by notificationsIndicatorViewModel.unreadCount.collectAsStateWithLifecycle()

    val currentBackStackEntry by navController.currentBackStackEntryAsState()
    val currentDestination = currentBackStackEntry?.destination
    val currentRoute = currentDestination?.route
    val selectedTab = currentDestination?.let(::tabOf)
    val isAtTabRoot = currentRoute != null && currentRoute == selectedTab?.startRoute

    LaunchedEffect(shouldOpenTransfers) {
        if (!shouldOpenTransfers) return@LaunchedEffect
        navController.openTransfers()
        onOpenTransfersHandled()
    }

    LaunchedEffect(requestedIngestUrl) {
        val sharedUrl = requestedIngestUrl ?: return@LaunchedEffect
        navController.openIngest(sharedUrl)
        onIngestUrlHandled()
    }

    LaunchedEffect(currentRoute) {
        if (currentRoute == AppRoute.MORE) notificationsIndicatorViewModel.refresh()
    }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { notificationsIndicatorViewModel.refresh() }

    BackHandler(enabled = shouldBackReturnToHome(selectedTab, isAtTabRoot)) {
        navController.openTab(BottomTab.HOME)
    }

    CompositionLocalProvider(LocalServerReconnections provides connectionBannerViewModel.reconnections) {
        Column(modifier = Modifier.fillMaxSize()) {
            Box(modifier = Modifier.weight(1f)) {
                NavHost(
                    navController = navController,
                    startDestination = BottomTab.HOME.graphRoute,
                    modifier = Modifier.fillMaxSize(),
                ) {
                    tabGraphs(navController)
                    topLevelRoutes(navController)
                }
                if (isBottomBarVisible(currentRoute)) {
                    MiniPlayer(
                        onExpand = { navController.navigate(AppRoute.MUSIC_PLAYER) },
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(horizontal = 12.dp, vertical = 12.dp),
                    )
                }
            }
            if (connectionState == ServerConnectionState.Offline) {
                ConnectionBanner(
                    onRetryNow = connectionBannerViewModel::retryNow,
                    onChangeServer = connectionBannerViewModel::changeServer,
                    modifier = if (isBottomBarVisible(currentRoute)) Modifier else Modifier.navigationBarsPadding(),
                )
            }
            if (isBottomBarVisible(currentRoute)) {
                AppBottomBar(
                    highlightedTab = highlightedTabFor(currentRoute, selectedTab),
                    isAtTabRoot = isAtTabRoot,
                    activeTransfersCount = activeTransfersCount,
                    unreadNotificationsCount = unreadNotificationsCount,
                    navController = navController,
                )
            }
        }
    }
}

@Composable
private fun AppBottomBar(
    highlightedTab: BottomTab?,
    isAtTabRoot: Boolean,
    activeTransfersCount: Int,
    unreadNotificationsCount: Int,
    navController: NavHostController,
) {
    NavigationBar(containerColor = BgCanvas) {
        BottomTab.entries.forEach { tab ->
            val label = stringResource(tab.labelRes)
            val isSelected = tab == highlightedTab
            NavigationBarItem(
                selected = isSelected,
                onClick = {
                    if (isSelected && !isAtTabRoot) {
                        navController.popBackStack(tab.startRoute, inclusive = false)
                        return@NavigationBarItem
                    }
                    navController.openTab(tab)
                },
                icon = {
                    BadgedBox(
                        badge = {
                            if (hasBadge(tab, activeTransfersCount, unreadNotificationsCount)) Badge()
                        },
                    ) {
                        Icon(tab.icon, contentDescription = null)
                    }
                },
                label = { Text(label) },
            )
        }
    }
}

private fun tabOf(destination: NavDestination): BottomTab? =
    BottomTab.entries.firstOrNull { tab -> destination.hierarchy.any { it.route == tab.graphRoute } }
