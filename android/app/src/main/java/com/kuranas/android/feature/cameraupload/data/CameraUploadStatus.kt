package com.kuranas.android.feature.cameraupload.data

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.combine
import javax.inject.Inject
import javax.inject.Singleton

data class CameraUploadStatus(
    val pendingCount: Int,
    val failedCount: Int,
    val lastSyncAtMillis: Long?,
    val lastSyncOutcome: CameraSyncOutcome?,
)

@Singleton
class CameraUploadStatusObserver @Inject constructor(
    settingsStore: CameraUploadSettingsStore,
    uploadDao: CameraUploadDao,
) {
    val status: Flow<CameraUploadStatus> = combine(settingsStore.settings, uploadDao.observeCounts()) { settings, counts ->
        CameraUploadStatus(
            pendingCount = counts.pendingCount,
            failedCount = counts.failedCount,
            lastSyncAtMillis = settings.lastSyncAtMillis,
            lastSyncOutcome = settings.lastSyncOutcome,
        )
    }
}
