package com.kuranas.android.feature.transfers.data.notification

object TransferBatchNotificationIds {

    fun progress(batchId: String): Int = baseOf(batchId) * 2

    fun result(batchId: String): Int = baseOf(batchId) * 2 + 1

    private fun baseOf(batchId: String): Int = batchId.hashCode() and MAX_BASE
}

private const val MAX_BASE = 0x3FFFFFFF
