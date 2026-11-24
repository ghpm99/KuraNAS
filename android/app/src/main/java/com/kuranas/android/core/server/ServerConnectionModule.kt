package com.kuranas.android.core.server

import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object ServerConnectionModule {

    @Provides
    @Singleton
    fun provideServerConnectionMonitor(
        serverStore: ServerStore,
        healthChecker: ServerHealthChecker,
        networkChangeEvents: NetworkChangeEvents,
        accessDeniedEvents: AccessDeniedEvents,
    ): ServerConnectionMonitor = ServerConnectionMonitor(
        serverUrl = serverStore.serverUrl,
        networkChanges = networkChangeEvents.events,
        checkServer = healthChecker::check,
        scope = CoroutineScope(SupervisorJob() + Dispatchers.Default),
        accessDeniedEvents = accessDeniedEvents,
    )
}
