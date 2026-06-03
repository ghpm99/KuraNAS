package com.kuranas.android.feature.files.ui

import android.content.ContentValues
import android.content.Context
import android.os.Environment
import android.provider.MediaStore
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileContent
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import javax.inject.Inject

sealed interface FileViewerState {
    data object Loading : FileViewerState
    data class Content(val content: FileContent, val info: FileItemDto?) : FileViewerState
    data class Failed(val message: String) : FileViewerState
}

@HiltViewModel
class FileViewerViewModel @Inject constructor(
    private val repository: FilesRepository,
    @ApplicationContext private val context: Context,
) : ViewModel() {

    private val _state = MutableStateFlow<FileViewerState>(FileViewerState.Loading)
    val state: StateFlow<FileViewerState> = _state.asStateFlow()

    private val _downloading = MutableStateFlow(false)
    val downloading: StateFlow<Boolean> = _downloading.asStateFlow()

    private val _messages = MutableSharedFlow<String>()
    val messages: SharedFlow<String> = _messages.asSharedFlow()

    fun load(fileId: String) {
        viewModelScope.launch {
            _state.value = FileViewerState.Loading
            val contentResult = repository.getFileContent(fileId)
            val info = (repository.getFileInfo(fileId) as? AppResult.Success)?.data
            _state.value = when (contentResult) {
                is AppResult.Success -> FileViewerState.Content(contentResult.data, info)
                is AppResult.Error -> FileViewerState.Failed(contentResult.message)
            }
        }
    }

    fun download(fileId: String, fileName: String) {
        if (_downloading.value) return
        viewModelScope.launch {
            _downloading.value = true
            when (val result = repository.getFileBytes(fileId)) {
                is AppResult.Success -> {
                    val saved = runCatching {
                        saveToDownloads(
                            name = fileName.ifBlank { "arquivo_$fileId" },
                            mime = result.data.contentType,
                            bytes = result.data.bytes,
                        )
                    }
                    _messages.emit(
                        if (saved.isSuccess) "Salvo em Downloads/${fileName.ifBlank { "arquivo_$fileId" }}"
                        else "Falha ao salvar o arquivo",
                    )
                }
                is AppResult.Error -> _messages.emit("Falha no download: ${result.message}")
            }
            _downloading.value = false
        }
    }

    private suspend fun saveToDownloads(name: String, mime: String?, bytes: ByteArray) =
        withContext(Dispatchers.IO) {
            val resolver = context.contentResolver
            val values = ContentValues().apply {
                put(MediaStore.Downloads.DISPLAY_NAME, name)
                put(MediaStore.Downloads.MIME_TYPE, mime ?: "application/octet-stream")
                put(MediaStore.Downloads.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
                put(MediaStore.Downloads.IS_PENDING, 1)
            }
            val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values)
                ?: error("Não foi possível criar o arquivo em Downloads")
            resolver.openOutputStream(uri)?.use { it.write(bytes) }
                ?: error("Não foi possível escrever o arquivo")
            values.clear()
            values.put(MediaStore.Downloads.IS_PENDING, 0)
            resolver.update(uri, values, null, null)
        }
}
