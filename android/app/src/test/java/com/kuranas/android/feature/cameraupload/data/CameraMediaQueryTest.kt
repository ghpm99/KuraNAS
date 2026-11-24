package com.kuranas.android.feature.cameraupload.data

import org.junit.Assert.assertEquals
import org.junit.Test

class CameraMediaQueryTest {

    @Test
    fun filtersCameraFolderAndWatermark() {
        val query = cameraMediaQuery(addedAfterSeconds = 1234)
        assertEquals(listOf("DCIM/Camera/%", "1234"), query.selectionArguments.toList())
    }
}
