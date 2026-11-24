package com.kuranas.android.feature.cameraupload.data

import androidx.work.NetworkType
import org.junit.Assert.assertEquals
import org.junit.Test

class CameraSyncConstraintsTest {

    @Test
    fun wifiOnlyRequiresUnmeteredNetwork() {
        val constraints = cameraSyncConstraintsFor(CameraUploadSettings(onlyOnWifi = true, onlyWhileCharging = false))
        assertEquals(CameraSyncConstraints(NetworkType.UNMETERED, requiresCharging = false), constraints)
    }

    @Test
    fun anyNetworkWhileChargingOnly() {
        val constraints = cameraSyncConstraintsFor(CameraUploadSettings(onlyOnWifi = false, onlyWhileCharging = true))
        assertEquals(CameraSyncConstraints(NetworkType.CONNECTED, requiresCharging = true), constraints)
    }
}
