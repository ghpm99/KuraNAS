package com.kuranas.android.navigation

import androidx.annotation.StringRes
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Book
import androidx.compose.material.icons.filled.Checklist
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Description
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.SwapVert
import androidx.compose.ui.graphics.vector.ImageVector
import com.kuranas.android.R

enum class MoreEntry(val route: String, @StringRes val labelRes: Int, val icon: ImageVector) {
    DOCUMENTS(AppRoute.DOCUMENTS, R.string.nav_documents, Icons.Filled.Description),
    FAVORITES(AppRoute.FAVORITES, R.string.nav_favorites, Icons.Filled.Star),
    SEARCH(AppRoute.SEARCH, R.string.nav_search, Icons.Filled.Search),
    TRANSFERS(AppRoute.TRANSFERS, R.string.nav_transfers, Icons.Filled.SwapVert),
    TRASH(AppRoute.TRASH, R.string.nav_trash, Icons.Filled.Delete),
    NOTIFICATIONS(AppRoute.NOTIFICATIONS, R.string.nav_notifications, Icons.Filled.Notifications),
    JOBS(AppRoute.JOBS, R.string.nav_jobs, Icons.Filled.Checklist),
    DIARY(AppRoute.DIARY, R.string.nav_diary, Icons.Filled.Book),
    SETTINGS(AppRoute.SETTINGS, R.string.nav_settings, Icons.Filled.Settings),
}

fun badgeCountFor(entry: MoreEntry, activeTransfersCount: Int, unreadNotificationsCount: Int): Int = when (entry) {
    MoreEntry.TRANSFERS -> activeTransfersCount
    MoreEntry.NOTIFICATIONS -> unreadNotificationsCount
    else -> 0
}
