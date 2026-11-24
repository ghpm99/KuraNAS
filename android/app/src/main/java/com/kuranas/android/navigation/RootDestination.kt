package com.kuranas.android.navigation

import com.kuranas.android.core.server.AccessDenial

sealed interface RootDestination {
    data object Checking : RootDestination
    data object Connection : RootDestination
    data object App : RootDestination
    data class AccessDenied(val denial: AccessDenial) : RootDestination
}
