package com.kuranas.android.navigation

import org.junit.Assert.assertEquals
import org.junit.Test

class MoreEntryTest {

    @Test
    fun `more lists only existing screens in the specified order`() {
        assertEquals(
            listOf(
                AppRoute.DOCUMENTS,
                AppRoute.FAVORITES,
                AppRoute.SEARCH,
                AppRoute.TRANSFERS,
                AppRoute.TRASH,
                AppRoute.NOTIFICATIONS,
                AppRoute.JOBS,
                AppRoute.INGEST,
                AppRoute.DIARY,
                AppRoute.SETTINGS,
            ),
            MoreEntry.entries.map { it.route },
        )
    }

    @Test
    fun `transfers and notifications show their counts and the others show none`() {
        assertEquals(4, badgeCountFor(MoreEntry.TRANSFERS, activeTransfersCount = 4, unreadNotificationsCount = 7))
        assertEquals(7, badgeCountFor(MoreEntry.NOTIFICATIONS, activeTransfersCount = 4, unreadNotificationsCount = 7))
        assertEquals(0, badgeCountFor(MoreEntry.SETTINGS, activeTransfersCount = 4, unreadNotificationsCount = 7))
    }
}
