package com.kuranas.android.feature.transfers.data

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.TypeConverters
import com.kuranas.android.feature.cameraupload.data.CameraUploadConverters
import com.kuranas.android.feature.cameraupload.data.CameraUploadDao
import com.kuranas.android.feature.cameraupload.data.CameraUploadEntity

@Database(entities = [TransferEntity::class, CameraUploadEntity::class], version = 4, exportSchema = true)
@TypeConverters(TransferConverters::class, CameraUploadConverters::class)
abstract class KuranasDatabase : RoomDatabase() {
    abstract fun transferDao(): TransferDao
    abstract fun cameraUploadDao(): CameraUploadDao
}
