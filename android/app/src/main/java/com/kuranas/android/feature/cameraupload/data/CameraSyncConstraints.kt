package com.kuranas.android.feature.cameraupload.data

import androidx.work.NetworkType

data class CameraSyncConstraints(val networkType: NetworkType, val requiresCharging: Boolean)

fun cameraSyncConstraintsFor(settings: CameraUploadSettings): CameraSyncConstraints =
    CameraSyncConstraints(
        networkType = if (settings.onlyOnWifi) NetworkType.UNMETERED else NetworkType.CONNECTED,
        requiresCharging = settings.onlyWhileCharging,
    )
