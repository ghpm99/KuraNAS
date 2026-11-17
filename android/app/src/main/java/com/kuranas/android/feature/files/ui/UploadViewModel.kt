package com.kuranas.android.feature.files.ui

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.SelectedFile
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.data.UploadDestinationStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class UploadViewModel @Inject constructor(
    private val repository: FilesRepository,
    private val batchRunner: UploadBatchRunner,
    private val destinationStore: UploadDestinationStore,
) : ViewModel() {

    private val mutableState = MutableStateFlow(UploadUiState())
    val state: StateFlow<UploadUiState> = mutableState.asStateFlow()

    private var pendingSelection: List<SelectedFile> = emptyList()

    /** Recebe os arquivos de um compartilhamento: envia direto ao destino fixo, se houver, ou abre a escolha de destino. */
    fun receiveShared(uris: List<Uri>) {
        if (mutableState.value.phase != UploadPhase.Idle) return
        mutableState.update { it.copy(phase = UploadPhase.ResolvingShare) }
        viewModelScope.launch {
            pendingSelection = repository.describeSelection(uris)
            val alwaysSendDestination = destinationStore.alwaysSendDestination.first()
            if (alwaysSendDestination != null) {
                send(alwaysSendDestination)
                return@launch
            }
            val lastDestination = destinationStore.lastDestination.first()
            mutableState.update { it.copy(phase = UploadPhase.ChoosingDestination(lastDestination)) }
        }
    }

    fun confirmDestination(destination: UploadDestination, shouldAlwaysSendHere: Boolean) {
        viewModelScope.launch {
            if (shouldAlwaysSendHere) destinationStore.saveAlwaysSendDestination(destination)
            send(destination)
        }
    }

    /** Envia os arquivos escolhidos no seletor do explorer para a pasta aberta. */
    fun sendPicked(uris: List<Uri>, destination: UploadDestination) {
        if (mutableState.value.phase is UploadPhase.Sending) return
        viewModelScope.launch {
            pendingSelection = repository.describeSelection(uris)
            send(destination)
        }
    }

    /** Reenvia apenas os arquivos que falharam no último envio, para o mesmo destino. */
    fun retryFailed() {
        val finished = mutableState.value.phase as? UploadPhase.Finished ?: return
        pendingSelection = finished.summary.failures.map { it.file }
        viewModelScope.launch { send(finished.destination) }
    }

    fun dismissResult() {
        mutableState.update { it.copy(phase = UploadPhase.Idle) }
    }

    private suspend fun send(destination: UploadDestination) {
        destinationStore.saveLastDestination(destination)
        val summary = batchRunner.upload(pendingSelection, destination.folderId) { current, total, fileName ->
            mutableState.update { it.copy(phase = UploadPhase.Sending(current, total, fileName)) }
        }
        pendingSelection = summary.failures.map { it.file }
        mutableState.update { it.copy(phase = UploadPhase.Finished(summary, destination)) }
    }
}
