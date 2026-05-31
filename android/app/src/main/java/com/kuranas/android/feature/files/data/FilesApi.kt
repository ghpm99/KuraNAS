package com.kuranas.android.feature.files.data

import com.kuranas.android.core.network.PageDto
import com.kuranas.android.core.network.PaginationDto
import com.kuranas.android.core.network.mimeTypeForFormat
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import okhttp3.RequestBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
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

    @Multipart
    @POST("api/v1/files/upload")
    suspend fun uploadFile(
        @Part file: MultipartBody.Part,
        @Part("folder_id") folderId: RequestBody,
    ): FileItemDto

    @POST("api/v1/files/folder")
    suspend fun createFolder(@Body body: CreateFolderRequest): FileItemDto

    @POST("api/v1/files/rename")
    suspend fun renameFile(@Body body: RenameRequest): FileItemDto

    @POST("api/v1/files/move")
    suspend fun moveFile(@Body body: MoveRequest)

    @POST("api/v1/files/copy")
    suspend fun copyFile(@Body body: MoveRequest)

    @DELETE("api/v1/files/path")
    suspend fun deleteFile(@Query("path") path: String)

    @POST("api/v1/files/starred/{id}")
    suspend fun starFile(@Path("id") id: String)
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
    val files: List<FileItemDto> = emptyList(),
    val total: Int = 0,
    @SerialName("total_size") val totalSize: Long = 0,
    val pagination: PaginationDto = PaginationDto(),
)

@Serializable
data class CreateFolderRequest(val name: String, @SerialName("parent_id") val parentId: String? = null)

@Serializable
data class RenameRequest(val id: String, val name: String)

@Serializable
data class MoveRequest(val id: String, @SerialName("target_id") val targetId: String)
