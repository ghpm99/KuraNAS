package com.kuranas.android.feature.files.data

import android.content.Context
import com.kuranas.android.R
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.paging.FIRST_PAGE
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.paging.toListPage
import android.net.Uri
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.toRequestBody
import dagger.Module
import dagger.Provides
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class FilesRepository @Inject constructor(
    private val api: FilesApi,
    @ApplicationContext private val context: Context,
) {

    suspend fun getRootFiles(page: Int): AppResult<ListPage<FileItemDto>> = safeApiCall {
        api.getFolderChildren(ROOT_FOLDER_ID, page, EXPLORER_PAGE_SIZE).toListPage()
    }

    suspend fun getChildrenById(id: String, page: Int): AppResult<ListPage<FileItemDto>> = safeApiCall {
        api.getFolderChildren(id, page, EXPLORER_PAGE_SIZE).toListPage()
    }

    suspend fun getStarredFiles(
        page: Int,
        kind: StarredKind? = null,
        sort: StarredSort? = null,
    ): AppResult<ListPage<FileItemDto>> = safeApiCall {
        api.getStarredFiles(page, STARRED_PAGE_SIZE, kind?.queryValue, sort?.queryValue).toListPage()
    }

    suspend fun describeSelection(uris: List<Uri>): List<SelectedFile> = withContext(Dispatchers.IO) {
        uris.map { describeSelectedFile(context, context.contentResolver, it) }
    }

    suspend fun uploadFile(file: SelectedFile, targetFolderId: String?): AppResult<UploadResponseDto> = safeApiCall {
        val contentResolver = context.contentResolver
        val mediaType = contentResolver.getType(file.uri)?.toMediaTypeOrNull()
        val content = UriRequestBody(contentResolver, file.uri, mediaType, file.sizeBytes)
        api.uploadFile(
            files = MultipartBody.Part.createFormData(UPLOAD_FILES_PART, file.name, content),
            targetFolderId = (targetFolderId ?: ROOT_FOLDER_ID).toRequestBody(),
            onConflict = UPLOAD_CONFLICT_POLICY.toRequestBody(),
        )
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

    suspend fun setFileStarred(id: String, isStarred: Boolean): AppResult<StarredResponse> = safeApiCall {
        api.setFileStarred(id, SetStarredRequest(isStarred))
    }

    /**
     * Metadados de um único arquivo. O backend não tem endpoint de detalhe: /files/{id}
     * lista os itens do caminho do arquivo, que para um arquivo é ele próprio na
     * primeira posição, então pedimos uma página de um item.
     */
    suspend fun getFileInfo(id: String): AppResult<FileItemDto> = safeApiCall {
        val items = api.getItemById(id, FIRST_PAGE, FILE_INFO_PAGE_SIZE).items
        items.firstOrNull { it.id == id }
            ?: items.firstOrNull()
            ?: throw NoSuchElementException(context.getString(R.string.file_not_found))
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
                contentType = contentType.ifBlank { context.getString(R.string.file_content_type_binary) },
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
        const val STARRED_PAGE_SIZE = 50
        const val EXPLORER_PAGE_SIZE = 50
        const val FILE_INFO_PAGE_SIZE = 1
        const val UPLOAD_FILES_PART = "files"
        const val UPLOAD_CONFLICT_POLICY = "rename"
        const val ROOT_FOLDER_ID = "0"
    }
}

enum class StarredKind(val queryValue: String) {
    ALL("all"),
    FOLDERS("folders"),
    FILES("files"),
    MEDIA("media"),
}

enum class StarredSort(val queryValue: String) {
    NAME("name"),
    STARRED_AT("starred_at"),
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
