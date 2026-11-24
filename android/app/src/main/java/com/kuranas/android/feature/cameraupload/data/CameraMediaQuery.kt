package com.kuranas.android.feature.cameraupload.data

import android.provider.MediaStore.MediaColumns

class CameraMediaQuery(val selection: String, val selectionArguments: Array<String>)

fun cameraMediaQuery(addedAfterSeconds: Long): CameraMediaQuery =
    CameraMediaQuery(
        selection = "${MediaColumns.RELATIVE_PATH} LIKE ? AND ${MediaColumns.DATE_ADDED} > ?",
        selectionArguments = arrayOf(CAMERA_RELATIVE_PATH_PATTERN, addedAfterSeconds.toString()),
    )

private const val CAMERA_RELATIVE_PATH_PATTERN = "DCIM/Camera/%"
