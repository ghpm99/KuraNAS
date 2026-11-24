package com.kuranas.android.feature.cameraupload.data

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import kotlinx.coroutines.flow.Flow

data class CameraUploadCounts(val pendingCount: Int, val failedCount: Int)

@Dao
interface CameraUploadDao {

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun registerAll(uploads: List<CameraUploadEntity>)

    @Query("SELECT mediaStoreId, mediaKind FROM camera_uploads")
    suspend fun registeredKeys(): List<CameraMediaKey>

    @Query(
        "SELECT " +
            "COALESCE(SUM(CASE WHEN t.status IN ('QUEUED', 'RUNNING', 'WAITING_FOR_NETWORK') THEN 1 ELSE 0 END), 0) AS pendingCount, " +
            "COALESCE(SUM(CASE WHEN t.status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failedCount " +
            "FROM camera_uploads c INNER JOIN transfers t ON t.id = c.transferId",
    )
    fun observeCounts(): Flow<CameraUploadCounts>
}
