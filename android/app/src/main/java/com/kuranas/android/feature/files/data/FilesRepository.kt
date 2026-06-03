package com.kuranas.android.feature.files.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import dagger.Module
import dagger.Provides
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class FilesRepository @Inject constructor(private val api: FilesApi) {

    suspend fun getRootFiles(): AppResult<List<FileItemDto>> = safeApiCall {
        api.getRootFiles().items
    }

    suspend fun getChildrenById(id: String): AppResult<List<FileItemDto>> = safeApiCall {
        api.getChildrenById(id).items
    }

    suspend fun getImages(): AppResult<List<FileItemDto>> = safeApiCall {
        api.getImages().items
    }

    suspend fun createFolder(name: String, parentId: String?): AppResult<Unit> = safeApiCall {
        api.createFolder(CreateFolderRequest(name, parentId?.toIntOrNull()))
    }

    suspend fun renameFile(id: String, name: String): AppResult<Unit> = safeApiCall {
        api.renameFile(RenameRequest(id.toIntOrNull() ?: 0, name))
    }

    suspend fun deleteFile(id: String): AppResult<Unit> = safeApiCall {
        api.deleteFile(DeleteFileRequest(id.toIntOrNull() ?: 0))
    }

    suspend fun starFile(id: String): AppResult<Unit> = safeApiCall {
        api.starFile(id)
    }

    /**
     * Metadados de um único arquivo. O endpoint /files/{id} devolve os arquivos do
     * mesmo caminho; pra um arquivo isso é ele próprio — então filtramos pelo id.
     */
    suspend fun getFileInfo(id: String): AppResult<FileItemDto> = safeApiCall {
        val items = api.getChildrenById(id).items
        items.firstOrNull { it.id == id }
            ?: items.firstOrNull()
            ?: throw NoSuchElementException("Arquivo não encontrado")
    }

    /** Baixa os bytes brutos do arquivo (pra salvar no dispositivo). */
    suspend fun getFileBytes(id: String): AppResult<FileBytes> = safeApiCall {
        val body = api.getBlob(id)
        FileBytes(bytes = body.bytes(), contentType = body.contentType()?.toString())
    }

    /**
     * Baixa o conteúdo do arquivo (endpoint /blob) e decide se é texto exibível
     * ou binário. O backend marca o Content-Type pela extensão, mas isso não
     * garante que o conteúdo seja texto legível — por isso também inspecionamos
     * os bytes (presença de NUL / proporção de caracteres imprimíveis).
     */
    suspend fun getFileContent(id: String): AppResult<FileContent> = safeApiCall {
        val body = api.getBlob(id)
        val contentType = body.contentType()?.toString().orEmpty()
        val bytes = body.bytes()
        val declaredText = contentType.startsWith("text/") ||
            contentType.contains("json") ||
            contentType.contains("xml") ||
            contentType.contains("csv")
        if ((declaredText || contentType.isBlank()) && isMostlyPrintable(bytes)) {
            val truncated = bytes.size > MAX_TEXT_BYTES
            val shown = if (truncated) bytes.copyOf(MAX_TEXT_BYTES) else bytes
            FileContent.Text(
                content = String(shown, Charsets.UTF_8),
                truncated = truncated,
            )
        } else {
            FileContent.Unsupported(
                contentType = contentType.ifBlank { "binário" },
                size = bytes.size.toLong(),
            )
        }
    }

    private fun isMostlyPrintable(bytes: ByteArray): Boolean {
        if (bytes.isEmpty()) return true
        val sample = bytes.take(MAX_TEXT_BYTES)
        var printable = 0
        for (b in sample) {
            val v = b.toInt() and 0xFF
            if (v == 0) return false // NUL → quase certamente binário
            if (v >= 0x20 || v == 9 || v == 10 || v == 13) printable++
        }
        return printable.toDouble() / sample.size >= 0.85
    }

    private companion object {
        const val MAX_TEXT_BYTES = 256 * 1024
    }
}

sealed interface FileContent {
    data class Text(val content: String, val truncated: Boolean) : FileContent
    data class Unsupported(val contentType: String, val size: Long) : FileContent
}

data class FileBytes(val bytes: ByteArray, val contentType: String?)

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object FilesModule {
    @Provides
    @Singleton
    fun provideFilesApi(retrofit: Retrofit): FilesApi = retrofit.create(FilesApi::class.java)
}
