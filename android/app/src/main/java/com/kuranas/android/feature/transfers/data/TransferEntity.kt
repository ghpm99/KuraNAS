package com.kuranas.android.feature.transfers.data

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "transfers")
data class TransferEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val batchId: String,
    val direction: TransferDirection,
    val displayName: String,
    val sourceUri: String? = null,
    val stagedFilePath: String? = null,
    val targetFolderId: Long? = null,
    val targetFolderName: String,
    val relativePath: String? = null,
    @ColumnInfo(defaultValue = "0") val requiresUnmeteredNetwork: Boolean = false,
    val remoteFileId: Long? = null,
    val downloadPurpose: String? = null,
    val destinationUri: String? = null,
    val mimeType: String? = null,
    val totalBytes: Long = -1,
    val transferredBytes: Long = 0,
    val uploadId: String? = null,
    val status: TransferStatus = TransferStatus.QUEUED,
    val failureMessage: String? = null,
    val createdAt: Long,
    val updatedAt: Long,
)
