package com.kuranas.android.feature.transfers.data

import javax.inject.Inject

fun interface TransferClock {
    fun nowMillis(): Long
}

sealed interface TransferTransition {
    data class Applied(val transfer: TransferEntity) : TransferTransition
    data class Rejected(val transfer: TransferEntity, val attemptedFrom: TransferStatus) : TransferTransition
}

class TransferStateMachine @Inject constructor(private val clock: TransferClock) {

    fun start(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.QUEUED || transfer.status == TransferStatus.WAITING_FOR_NETWORK) {
            it.copy(status = TransferStatus.RUNNING)
        }

    fun progress(transfer: TransferEntity, transferredBytes: Long): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.RUNNING) {
            it.copy(transferredBytes = transferredBytes)
        }

    fun waitForNetwork(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.RUNNING) {
            it.copy(status = TransferStatus.WAITING_FOR_NETWORK)
        }

    fun complete(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.RUNNING) {
            it.copy(
                status = TransferStatus.COMPLETED,
                transferredBytes = if (it.totalBytes >= 0) it.totalBytes else it.transferredBytes,
            )
        }

    fun alreadyExisted(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.RUNNING) {
            it.copy(status = TransferStatus.ALREADY_EXISTED)
        }

    fun fail(transfer: TransferEntity, failureMessage: String): TransferTransition =
        transition(transfer, transfer.status.isActive) {
            it.copy(status = TransferStatus.FAILED, failureMessage = failureMessage)
        }

    fun cancel(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status.isActive) {
            it.copy(status = TransferStatus.CANCELLED)
        }

    fun retry(transfer: TransferEntity): TransferTransition =
        transition(transfer, transfer.status == TransferStatus.FAILED || transfer.status == TransferStatus.CANCELLED) {
            it.copy(status = TransferStatus.QUEUED, failureMessage = null)
        }

    private fun transition(
        transfer: TransferEntity,
        isAllowed: Boolean,
        apply: (TransferEntity) -> TransferEntity,
    ): TransferTransition {
        if (!isAllowed) return TransferTransition.Rejected(transfer, transfer.status)
        return TransferTransition.Applied(apply(transfer).copy(updatedAt = clock.nowMillis()))
    }
}
