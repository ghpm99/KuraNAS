package com.kuranas.android.feature.files.ui

import com.kuranas.android.feature.files.data.UploadDestination

sealed interface UploadPhase {
    data object Idle : UploadPhase
    data object ResolvingShare : UploadPhase
    data class ChoosingDestination(val lastDestination: UploadDestination?) : UploadPhase
    data object Preparing : UploadPhase
    data class Queued(val queuedCount: Int, val unreadableFileNames: List<String>) : UploadPhase
}

data class UploadUiState(val phase: UploadPhase = UploadPhase.Idle)
