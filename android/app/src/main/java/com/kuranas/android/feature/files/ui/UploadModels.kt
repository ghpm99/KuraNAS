package com.kuranas.android.feature.files.ui

import com.kuranas.android.feature.files.data.SelectedFile
import com.kuranas.android.feature.files.data.UploadDestination

sealed interface UploadFailureReason {
    data object NoServer : UploadFailureReason
    data class ServerMessage(val text: String) : UploadFailureReason
}

data class UploadFailure(val file: SelectedFile, val reason: UploadFailureReason)

data class UploadSummary(
    val uploadedCount: Int = 0,
    val renamedCount: Int = 0,
    val skippedCount: Int = 0,
    val failures: List<UploadFailure> = emptyList(),
) {
    val hasFailures: Boolean get() = failures.isNotEmpty()
}

sealed interface UploadPhase {
    data object Idle : UploadPhase
    data object ResolvingShare : UploadPhase
    data class ChoosingDestination(val lastDestination: UploadDestination?) : UploadPhase
    data class Sending(val current: Int, val total: Int, val fileName: String) : UploadPhase
    data class Finished(val summary: UploadSummary, val destination: UploadDestination) : UploadPhase
}

data class UploadUiState(val phase: UploadPhase = UploadPhase.Idle)
