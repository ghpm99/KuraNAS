package com.kuranas.android.feature.transfers.data

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Transaction
import androidx.room.Update
import kotlinx.coroutines.flow.Flow

@Dao
abstract class TransferDao {

    @Insert
    abstract suspend fun insertAll(transfers: List<TransferEntity>): List<Long>

    @Query("SELECT * FROM transfers ORDER BY createdAt DESC, id DESC")
    abstract fun observeAll(): Flow<List<TransferEntity>>

    @Query("SELECT COUNT(*) FROM transfers WHERE status IN ('QUEUED', 'RUNNING', 'WAITING_FOR_NETWORK')")
    abstract fun observeActiveCount(): Flow<Int>

    @Query("SELECT * FROM transfers WHERE id = :transferId")
    abstract suspend fun findById(transferId: Long): TransferEntity?

    @Query("SELECT * FROM transfers WHERE batchId = :batchId ORDER BY id ASC")
    abstract suspend fun findByBatch(batchId: String): List<TransferEntity>

    @Update
    abstract suspend fun update(transfer: TransferEntity)

    @Query("DELETE FROM transfers WHERE status IN ('COMPLETED', 'ALREADY_EXISTED', 'CANCELLED')")
    abstract suspend fun deleteFinished()

    @Transaction
    open suspend fun applyTransition(transferId: Long, transition: (TransferEntity) -> TransferTransition): TransferEntity? {
        val current = findById(transferId) ?: return null
        val applied = transition(current) as? TransferTransition.Applied ?: return null
        update(applied.transfer)
        return applied.transfer
    }
}
