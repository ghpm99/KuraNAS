package com.kuranas.android.navigation

private val fullScreenRoutes = setOf(
    AppRoute.MUSIC_PLAYER_ROUTE,
    AppRoute.VIDEO_PLAYER,
    AppRoute.IMAGE_VIEWER,
    AppRoute.FILE_DETAIL,
)

fun isBottomBarVisible(route: String?): Boolean = route !in fullScreenRoutes

fun shouldBackReturnToHome(selectedTab: BottomTab?, isAtTabRoot: Boolean): Boolean =
    isAtTabRoot && selectedTab != null && selectedTab != BottomTab.HOME

fun hasBadge(tab: BottomTab, activeTransfersCount: Int, unreadNotificationsCount: Int): Boolean =
    tab == BottomTab.MORE && (activeTransfersCount > 0 || unreadNotificationsCount > 0)
