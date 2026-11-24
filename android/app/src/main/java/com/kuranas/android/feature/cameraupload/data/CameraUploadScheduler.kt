package com.kuranas.android.feature.cameraupload.data

import android.provider.MediaStore
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.OutOfQuotaPolicy
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CameraUploadScheduler @Inject constructor(private val workManager: WorkManager) {

    fun apply(settings: CameraUploadSettings) {
        if (!settings.isEnabled) {
            cancelAll()
            return
        }
        schedulePeriodic(settings)
        scheduleChangeTrigger(settings, ExistingWorkPolicy.REPLACE)
    }

    fun scheduleChangeTrigger(settings: CameraUploadSettings, existingWorkPolicy: ExistingWorkPolicy) {
        val constraints = constraintsOf(settings).apply {
            setTriggerContentUpdateDelay(TRIGGER_UPDATE_DELAY_SECONDS, TimeUnit.SECONDS)
            setTriggerContentMaxDelay(TRIGGER_MAX_DELAY_SECONDS, TimeUnit.SECONDS)
            addContentUriTrigger(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, true)
            addContentUriTrigger(MediaStore.Video.Media.EXTERNAL_CONTENT_URI, true)
        }
        val request = OneTimeWorkRequestBuilder<CameraSyncWorker>()
            .setConstraints(constraints.build())
            .build()
        workManager.enqueueUniqueWork(CHANGE_TRIGGER_WORK_NAME, existingWorkPolicy, request)
    }

    fun syncNow() {
        val request = OneTimeWorkRequestBuilder<CameraSyncWorker>()
            .setConstraints(Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build())
            .setExpedited(OutOfQuotaPolicy.RUN_AS_NON_EXPEDITED_WORK_REQUEST)
            .build()
        workManager.enqueueUniqueWork(SYNC_NOW_WORK_NAME, ExistingWorkPolicy.KEEP, request)
    }

    private fun schedulePeriodic(settings: CameraUploadSettings) {
        val request = PeriodicWorkRequestBuilder<CameraSyncWorker>(PERIODIC_INTERVAL_MINUTES, TimeUnit.MINUTES)
            .setConstraints(constraintsOf(settings).build())
            .build()
        workManager.enqueueUniquePeriodicWork(PERIODIC_WORK_NAME, ExistingPeriodicWorkPolicy.UPDATE, request)
    }

    private fun cancelAll() {
        workManager.cancelUniqueWork(PERIODIC_WORK_NAME)
        workManager.cancelUniqueWork(CHANGE_TRIGGER_WORK_NAME)
        workManager.cancelUniqueWork(SYNC_NOW_WORK_NAME)
    }

    private fun constraintsOf(settings: CameraUploadSettings): Constraints.Builder {
        val syncConstraints = cameraSyncConstraintsFor(settings)
        return Constraints.Builder()
            .setRequiredNetworkType(syncConstraints.networkType)
            .setRequiresCharging(syncConstraints.requiresCharging)
    }

    private companion object {
        const val PERIODIC_WORK_NAME = "camera-upload-periodic"
        const val CHANGE_TRIGGER_WORK_NAME = "camera-upload-change-trigger"
        const val SYNC_NOW_WORK_NAME = "camera-upload-sync-now"
        const val PERIODIC_INTERVAL_MINUTES = 15L
        const val TRIGGER_UPDATE_DELAY_SECONDS = 5L
        const val TRIGGER_MAX_DELAY_SECONDS = 30L
    }
}
