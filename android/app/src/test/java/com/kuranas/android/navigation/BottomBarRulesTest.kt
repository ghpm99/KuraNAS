package com.kuranas.android.navigation

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class BottomBarRulesTest {

    @Test
    fun `bar is hidden on full screen routes`() {
        listOf(AppRoute.MUSIC_PLAYER_ROUTE, AppRoute.VIDEO_PLAYER, AppRoute.IMAGE_VIEWER, AppRoute.FILE_DETAIL)
            .forEach { route -> assertFalse(route, isBottomBarVisible(route)) }
    }

    @Test
    fun `bar is shown on tab roots, inner routes and secondary screens`() {
        listOf(AppRoute.HOME, AppRoute.FILES, AppRoute.MUSIC_ALBUM, AppRoute.VIDEO_LIBRARY, AppRoute.MORE, AppRoute.TRANSFERS, null)
            .forEach { route -> assertTrue(route.toString(), isBottomBarVisible(route)) }
    }

    @Test
    fun `back at the root of a tab other than home goes to home`() {
        assertTrue(shouldBackReturnToHome(BottomTab.FILES, isAtTabRoot = true))
        assertTrue(shouldBackReturnToHome(BottomTab.MORE, isAtTabRoot = true))
    }

    @Test
    fun `back at home root or inside a tab stack does not go to home`() {
        assertFalse(shouldBackReturnToHome(BottomTab.HOME, isAtTabRoot = true))
        assertFalse(shouldBackReturnToHome(BottomTab.MUSIC, isAtTabRoot = false))
        assertFalse(shouldBackReturnToHome(null, isAtTabRoot = true))
    }

    @Test
    fun `only the more tab carries a badge and only when something is pending`() {
        assertTrue(hasBadge(BottomTab.MORE, activeTransfersCount = 1, unreadNotificationsCount = 0))
        assertTrue(hasBadge(BottomTab.MORE, activeTransfersCount = 0, unreadNotificationsCount = 3))
        assertFalse(hasBadge(BottomTab.MORE, 0, 0))
        assertFalse(hasBadge(BottomTab.FILES, 2, 2))
    }

    @Test
    fun `tabs are home, files, images, music, video and more in that order`() {
        assertEquals(
            listOf(BottomTab.HOME, BottomTab.FILES, BottomTab.IMAGES, BottomTab.MUSIC, BottomTab.VIDEO, BottomTab.MORE),
            BottomTab.entries.toList(),
        )
    }

    @Test
    fun `search highlights no tab and keeps the bar visible`() {
        assertEquals(null, highlightedTabFor(AppRoute.SEARCH, null))
        assertTrue(isBottomBarVisible(AppRoute.SEARCH))
    }

    @Test
    fun `tab routes highlight their tab and other top level routes fall back to more`() {
        assertEquals(BottomTab.FILES, highlightedTabFor(AppRoute.FILES, BottomTab.FILES))
        assertEquals(BottomTab.MORE, highlightedTabFor(AppRoute.TRANSFERS, null))
    }
}
