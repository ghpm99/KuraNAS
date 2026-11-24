package com.kuranas.android.feature.transfers.data

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Update
import kotlinx.coroutines.flow.Flow

@Dao
interface TransferDao {

    @Insert
    suspend fun insertAll(transfers: List<TransferEntity>): List<Long>

    @Query("SELECT * FROM transfers ORDER BY createdAt DESC, id DESC")
    fun observeAll(): Flow<List<TransferEntity>>

    @Query("SELECT COUNT(*) FROM transfers WHERE status IN ('QUEUED', 'RUNNING', 'WAITING_FOR_NETWORK')")
    fun observeActiveCount(): Flow<Int>

    @Query("SELECT * FROM transfers WHERE id = :transferId")
    suspend fun findById(transferId: Long): TransferEntity?

    @Query("SELECT * FROM transfers WHERE batchId = :batchId ORDER BY id ASC")
    suspend fun findByBatch(batchId: String): List<TransferEntity>

    @Update
    suspend fun update(transfer: TransferEntity)

    @Query("DELETE FROM transfers WHERE status IN ('COMPLETED', 'ALREADY_EXISTED', 'CANCELLED')")
    suspend fun deleteFinished()
}
