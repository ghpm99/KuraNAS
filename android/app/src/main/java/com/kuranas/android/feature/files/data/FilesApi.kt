package com.kuranas.android.feature.files.data

import com.kuranas.android.core.network.PageDto
import com.kuranas.android.core.network.PaginationDto
import com.kuranas.android.core.network.mimeTypeForFormat
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import okhttp3.RequestBody
import okhttp3.ResponseBody
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.HTTP
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path
import retrofit2.http.Query

interface FilesApi {
    @GET("api/v1/files/")
    suspend fun getRootFiles(): PageDto<FileItemDto>

    @GET("api/v1/files/{id}")
    suspend fun getChildrenById(@Path("id") id: String): PageDto<FileItemDto>

    @GET("api/v1/files/path")
    suspend fun getFilesByPath(@Query("path") path: String): PageDto<FileItemDto>

    @GET("api/v1/files/images")
    suspend fun getImages(@Query("page") page: Int = 1, @Query("limit") limit: Int = 50): PageDto<FileItemDto>

    @GET("api/v1/files/music")
    suspend fun getMusicFiles(): PageDto<FileItemDto>

    @GET("api/v1/files/videos")
    suspend fun getVideoFiles(): PageDto<FileItemDto>

    @GET("api/v1/files/duplicate-files")
    suspend fun getDuplicateFiles(): DuplicateFilesDto

    // O part do arquivo deve se chamar `files` e o id da pasta destino é
    // `target_folder_id`; o handler responde 202 com { message, uploaded, job_id }.
    @Multipart
    @POST("api/v1/files/upload")
    suspend fun uploadFile(
        @Part file: MultipartBody.Part,
        @Part("target_folder_id") targetFolderId: RequestBody,
    )

    @POST("api/v1/files/folder")
    suspend fun createFolder(@Body body: CreateFolderRequest)

    @POST("api/v1/files/rename")
    suspend fun renameFile(@Body body: RenameRequest)

    @POST("api/v1/files/move")
    suspend fun moveFile(@Body body: MoveRequest)

    @POST("api/v1/files/copy")
    suspend fun copyFile(@Body body: CopyRequest)

    // DELETE com corpo JSON { id }, como o handler espera (não query param).
    @HTTP(method = "DELETE", path = "api/v1/files/path", hasBody = true)
    suspend fun deleteFile(@Body body: DeleteFileRequest)

    @POST("api/v1/files/starred/{id}")
    suspend fun starFile(@Path("id") id: String)

    @GET("api/v1/files/blob/{id}")
    suspend fun getBlob(@Path("id") id: String): ResponseBody
}

/**
 * Espelha files.FileDto do backend Go. As propriedades computadas (id/isDir/
 * isStarred/mimeType) mantêm a API esperada pela UI a partir dos campos reais.
 */
@Serializable
data class FileItemDto(
    @SerialName("id") val rawId: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val type: Int = TYPE_FILE,
    val format: String = "",
    val size: Long = 0,
    val starred: Boolean = false,
    @SerialName("directory_content_count") val directoryContentCount: Int = 0,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
) {
    val id: String get() = rawId.toString()
    val isDir: Boolean get() = type == TYPE_DIRECTORY
    val isStarred: Boolean get() = starred
    val mimeType: String get() = mimeTypeForFormat(format)

    companion object {
        const val TYPE_DIRECTORY = 1
        const val TYPE_FILE = 2
    }
}

@Serializable
data class DuplicateFilesDto(
    val files: List<DuplicateFileDto> = emptyList(),
    val total: Int = 0,
    @SerialName("total_size") val totalSize: Long = 0,
    val pagination: PaginationDto = PaginationDto(),
)

/** Espelha files.DuplicateFileDto: um grupo de cópias do mesmo arquivo. */
@Serializable
data class DuplicateFileDto(
    val name: String = "",
    val size: Long = 0,
    val copies: Int = 0,
    val paths: List<String> = emptyList(),
)

@Serializable
data class CreateFolderRequest(val name: String, @SerialName("parent_id") val parentId: Int? = null)

@Serializable
data class RenameRequest(val id: Int, @SerialName("new_name") val newName: String)

@Serializable
data class MoveRequest(
    @SerialName("source_id") val sourceId: Int,
    @SerialName("destination_folder_id") val destinationFolderId: Int? = null,
    @SerialName("destination_path") val destinationPath: String = "",
)

@Serializable
data class CopyRequest(
    @SerialName("source_id") val sourceId: Int,
    @SerialName("destination_folder_id") val destinationFolderId: Int? = null,
    @SerialName("destination_path") val destinationPath: String = "",
    @SerialName("new_name") val newName: String = "",
)

@Serializable
data class DeleteFileRequest(val id: Int)
