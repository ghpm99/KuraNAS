package com.kuranas.android.feature.cameraupload.data

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CameraUploadStartup @Inject constructor(
    private val settingsStore: CameraUploadSettingsStore,
    private val scheduler: CameraUploadScheduler,
) {

    private val startupScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    fun reapplySchedule() {
        startupScope.launch { scheduler.apply(settingsStore.current()) }
    }
}
