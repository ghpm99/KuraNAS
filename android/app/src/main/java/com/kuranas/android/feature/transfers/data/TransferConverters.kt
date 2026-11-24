package com.kuranas.android.feature.transfers.data

import androidx.room.TypeConverter

class TransferConverters {

    @TypeConverter
    fun directionToName(direction: TransferDirection): String = direction.name

    @TypeConverter
    fun directionFromName(name: String): TransferDirection = TransferDirection.valueOf(name)

    @TypeConverter
    fun statusToName(status: TransferStatus): String = status.name

    @TypeConverter
    fun statusFromName(name: String): TransferStatus = TransferStatus.valueOf(name)
}
