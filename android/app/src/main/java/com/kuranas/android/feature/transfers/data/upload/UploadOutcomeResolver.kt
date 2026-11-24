package com.kuranas.android.feature.transfers.data.upload

import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStateMachine
import com.kuranas.android.feature.transfers.data.TransferTransition

data class UploadFailureMessages(val sourceMissing: String, val rejectedWithoutDetail: String)

data class UploadOutcomeResolution(val transition: TransferTransition, val shouldRetryLater: Boolean)

class UploadOutcomeResolver(
    private val stateMachine: TransferStateMachine,
    private val failureMessages: UploadFailureMessages,
) {

    fun resolve(transfer: TransferEntity, outcome: UploadOutcome): UploadOutcomeResolution =
        when (outcome) {
            UploadOutcome.Completed -> UploadOutcomeResolution(stateMachine.complete(transfer), false)
            UploadOutcome.AlreadyExisted -> UploadOutcomeResolution(stateMachine.alreadyExisted(transfer), false)
            UploadOutcome.ConnectionLost -> UploadOutcomeResolution(stateMachine.waitForNetwork(transfer), true)
            UploadOutcome.SourceMissing ->
                UploadOutcomeResolution(stateMachine.fail(transfer, failureMessages.sourceMissing), false)
            is UploadOutcome.Rejected -> UploadOutcomeResolution(
                stateMachine.fail(transfer, outcome.serverMessage ?: failureMessages.rejectedWithoutDetail),
                false,
            )
        }
}
