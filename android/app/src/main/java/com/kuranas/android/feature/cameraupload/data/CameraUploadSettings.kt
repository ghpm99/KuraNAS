package com.kuranas.android.feature.cameraupload.data

import com.kuranas.android.feature.files.data.UploadDestination

enum class CameraSyncOutcome {
    DISABLED,
    AWAY_FROM_HOME,
    MEDIA_PERMISSION_MISSING,
    UP_TO_DATE,
    ENQUEUED,
}

data class CameraUploadSettings(
    val isEnabled: Boolean = false,
    val destination: UploadDestination = UploadDestination.Root,
    val relativePrefix: String = DEFAULT_RELATIVE_PREFIX,
    val onlyOnWifi: Boolean = true,
    val onlyWhileCharging: Boolean = false,
    val includeVideos: Boolean = true,
    val startingPointSeconds: Long = 0,
    val lastSyncAtMillis: Long? = null,
    val lastSyncOutcome: CameraSyncOutcome? = null,
) {
    companion object {
        const val DEFAULT_RELATIVE_PREFIX = "Fotos/Celular"
    }
}
