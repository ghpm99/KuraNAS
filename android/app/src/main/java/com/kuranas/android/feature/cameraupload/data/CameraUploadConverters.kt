package com.kuranas.android.feature.cameraupload.data

import androidx.room.TypeConverter

class CameraUploadConverters {

    @TypeConverter
    fun mediaKindToName(kind: CameraMediaKind): String = kind.name

    @TypeConverter
    fun mediaKindFromName(name: String): CameraMediaKind = CameraMediaKind.valueOf(name)
}
