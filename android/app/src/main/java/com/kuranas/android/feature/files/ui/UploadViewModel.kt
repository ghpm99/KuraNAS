package com.kuranas.android.feature.files.ui

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.data.UploadDestinationStore
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import com.kuranas.android.feature.transfers.data.upload.UploadRequestItem
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.io.IOException
import javax.inject.Inject

@HiltViewModel
class UploadViewModel @Inject constructor(
    private val transferRepository: TransferRepository,
    private val destinationStore: UploadDestinationStore,
) : ViewModel() {

    private val mutableState = MutableStateFlow(UploadUiState())
    val state: StateFlow<UploadUiState> = mutableState.asStateFlow()

    private var pendingSharedUris: List<Uri> = emptyList()

    /** Recebe os arquivos de um compartilhamento: enfileira direto no destino fixo, se houver, ou abre a escolha de destino. */
    fun receiveShared(uris: List<Uri>) {
        if (mutableState.value.phase != UploadPhase.Idle) return
        mutableState.update { it.copy(phase = UploadPhase.ResolvingShare) }
        pendingSharedUris = uris
        viewModelScope.launch {
            val alwaysSendDestination = destinationStore.alwaysSendDestination.first()
            if (alwaysSendDestination != null) {
                enqueueShared(alwaysSendDestination)
                return@launch
            }
            val lastDestination = destinationStore.lastDestination.first()
            mutableState.update { it.copy(phase = UploadPhase.ChoosingDestination(lastDestination)) }
        }
    }

    fun confirmDestination(destination: UploadDestination, shouldAlwaysSendHere: Boolean) {
        viewModelScope.launch {
            if (shouldAlwaysSendHere) destinationStore.saveAlwaysSendDestination(destination)
            enqueueShared(destination)
        }
    }

    /** Enfileira os arquivos escolhidos no seletor do explorer para a pasta aberta. */
    fun sendPicked(uris: List<Uri>, destination: UploadDestination) {
        if (mutableState.value.phase is UploadPhase.Preparing) return
        mutableState.update { it.copy(phase = UploadPhase.Preparing) }
        viewModelScope.launch {
            enqueue(uris, destination) { uri -> transferRepository.persistPickedUri(uri) }
        }
    }

    fun dismissResult() {
        mutableState.update { it.copy(phase = UploadPhase.Idle) }
    }

    private suspend fun enqueueShared(destination: UploadDestination) {
        mutableState.update { it.copy(phase = UploadPhase.Preparing) }
        enqueue(pendingSharedUris, destination) { uri -> stagedItemOf(uri) }
    }

    private suspend fun stagedItemOf(uri: Uri): UploadRequestItem {
        val stagedPath = transferRepository.stageSharedUri(uri)
        return withContext(Dispatchers.IO) {
            val stagedFile = File(stagedPath)
            UploadRequestItem(stagedFile.name, stagedFile.length(), stagedFilePath = stagedPath)
        }
    }

    private suspend fun enqueue(
        uris: List<Uri>,
        destination: UploadDestination,
        prepareItem: suspend (Uri) -> UploadRequestItem,
    ) {
        destinationStore.saveLastDestination(destination)
        val preparedItems = mutableListOf<UploadRequestItem>()
        val unreadableFileNames = mutableListOf<String>()
        uris.forEach { uri ->
            try {
                preparedItems += prepareItem(uri)
            } catch (exception: IOException) {
                unreadableFileNames += uri.lastPathSegment.orEmpty()
            } catch (exception: SecurityException) {
                unreadableFileNames += uri.lastPathSegment.orEmpty()
            }
        }
        if (preparedItems.isNotEmpty()) transferRepository.enqueueUploads(preparedItems, destination)
        pendingSharedUris = emptyList()
        mutableState.update { it.copy(phase = UploadPhase.Queued(preparedItems.size, unreadableFileNames)) }
    }
}
