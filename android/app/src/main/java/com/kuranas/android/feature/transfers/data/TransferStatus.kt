package com.kuranas.android.feature.transfers.data

enum class TransferStatus {
    QUEUED,
    RUNNING,
    WAITING_FOR_NETWORK,
    COMPLETED,
    ALREADY_EXISTED,
    FAILED,
    CANCELLED;

    val isActive: Boolean
        get() = this == QUEUED || this == RUNNING || this == WAITING_FOR_NETWORK
}
