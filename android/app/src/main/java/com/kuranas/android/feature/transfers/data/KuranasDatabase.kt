package com.kuranas.android.feature.transfers.data

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.TypeConverters

@Database(entities = [TransferEntity::class], version = 1, exportSchema = true)
@TypeConverters(TransferConverters::class)
abstract class KuranasDatabase : RoomDatabase() {
    abstract fun transferDao(): TransferDao
}
