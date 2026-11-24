package com.kuranas.android.feature.transfers.data.download

import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferTransition

data class DownloadFailureMessages(
    val storageUnavailable: String,
    val serverRejectedWithoutDetail: (httpCode: Int) -> String,
)

data class DownloadOutcomeResolution(val transition: TransferTransition, val shouldRetryLater: Boolean)

class DownloadOutcomeResolver(
    private val stateMachine: TransferStateMachine,
    private val failureMessages: DownloadFailureMessages,
) {

    fun resolve(transfer: TransferEntity, outcome: DownloadOutcome): DownloadOutcomeResolution =
        when (outcome) {
            is DownloadOutcome.Completed -> DownloadOutcomeResolution(completed(transfer, outcome), false)
            is DownloadOutcome.Failed -> resolveFailure(transfer, outcome.failure)
        }

    private fun resolveFailure(transfer: TransferEntity, failure: DownloadFailure): DownloadOutcomeResolution =
        when (failure) {
            DownloadFailure.ConnectionLost -> DownloadOutcomeResolution(stateMachine.waitForNetwork(transfer), true)
            DownloadFailure.StorageUnavailable ->
                DownloadOutcomeResolution(failed(transfer, failureMessages.storageUnavailable), false)
            is DownloadFailure.Server -> DownloadOutcomeResolution(
                failed(transfer, failure.message ?: failureMessages.serverRejectedWithoutDetail(failure.httpCode)),
                false,
            )
        }

    private fun completed(transfer: TransferEntity, outcome: DownloadOutcome.Completed): TransferTransition =
        when (val transition = stateMachine.complete(transfer)) {
            is TransferTransition.Applied -> TransferTransition.Applied(
                transition.transfer.copy(destinationUri = outcome.uri.toString(), mimeType = outcome.mimeType),
            )
            is TransferTransition.Rejected -> transition
        }

    private fun failed(transfer: TransferEntity, failureMessage: String): TransferTransition =
        when (val transition = stateMachine.fail(transfer, failureMessage)) {
            is TransferTransition.Applied -> TransferTransition.Applied(transition.transfer.withoutPartialDownload())
            is TransferTransition.Rejected -> transition
        }
}

fun TransferEntity.withoutPartialDownload(): TransferEntity =
    copy(destinationUri = null, mimeType = null, transferredBytes = 0)
