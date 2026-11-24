package com.kuranas.android.feature.cameraupload.data

data class CameraMediaKey(val mediaStoreId: Long, val mediaKind: CameraMediaKind)

data class CameraMedia(
    val id: Long,
    val kind: CameraMediaKind,
    val contentUri: String,
    val displayName: String,
    val sizeBytes: Long,
    val dateTakenMillis: Long,
    val dateModifiedMillis: Long,
) {
    val key: CameraMediaKey get() = CameraMediaKey(id, kind)
}
