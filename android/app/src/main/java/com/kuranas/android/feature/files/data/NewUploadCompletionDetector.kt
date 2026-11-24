package com.kuranas.android.feature.files.data

import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus

/**
 * Tracks which uploads were already completed so only completions observed after the first
 * snapshot count. Root is represented by a null folder id.
 */
class NewUploadCompletionDetector {

    private var alreadyCompletedIds: Set<Long>? = null

    /**
     * Records the completed uploads of [transfers] and returns whether any of them is new
     * and targets [openFolderId]. The first snapshot only establishes the baseline.
     */
    fun hasNewCompletionIn(transfers: List<TransferEntity>, openFolderId: Long?): Boolean {
        val completedUploads = transfers.filter { it.direction == TransferDirection.UPLOAD && it.status == TransferStatus.COMPLETED }
        val previouslyCompletedIds = alreadyCompletedIds
        alreadyCompletedIds = completedUploads.mapTo(mutableSetOf()) { it.id }
        if (previouslyCompletedIds == null) return false
        return completedUploads.any { it.id !in previouslyCompletedIds && it.targetFolderId == openFolderId }
    }
}
