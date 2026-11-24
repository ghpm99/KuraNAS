package com.kuranas.android.feature.cameraupload.data

import androidx.room.Entity
import androidx.room.Index

@Entity(
    tableName = "camera_uploads",
    primaryKeys = ["mediaStoreId", "mediaKind"],
    indices = [Index("transferId")],
)
data class CameraUploadEntity(
    val mediaStoreId: Long,
    val mediaKind: CameraMediaKind,
    val transferId: Long,
    val displayName: String,
    val dateTakenMillis: Long,
    val registeredAt: Long,
)
