package com.kuranas.android.feature.cameraupload.data

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.ExistingWorkPolicy
import androidx.work.WorkerParameters
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject

@HiltWorker
class CameraSyncWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParameters: WorkerParameters,
    private val runner: CameraSyncRunner,
    private val settingsStore: CameraUploadSettingsStore,
    private val scheduler: CameraUploadScheduler,
) : CoroutineWorker(appContext, workerParameters) {

    override suspend fun doWork(): Result {
        try {
            runner.run()
        } finally {
            val settings = settingsStore.current()
            if (settings.isEnabled) scheduler.scheduleChangeTrigger(settings, ExistingWorkPolicy.APPEND_OR_REPLACE)
        }
        return Result.success()
    }
}
