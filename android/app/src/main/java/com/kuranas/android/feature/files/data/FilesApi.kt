package com.kuranas.android.feature.files.data

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
    suspend fun getRootFiles(): FilesResponseDto

    @GET("api/v1/files/{id}")
    suspend fun getChildrenById(@Path("id") id: String): FilesResponseDto

    @GET("api/v1/files/path")
    suspend fun getFilesByPath(@Query("path") path: String): FilesResponseDto

    @GET("api/v1/files/recent")
    suspend fun getRecentFiles(): List<FileItemDto>

    @GET("api/v1/files/images")
    suspend fun getImages(@Query("page") page: Int = 1, @Query("limit") limit: Int = 50): FilesResponseDto

    @GET("api/v1/files/music")
    suspend fun getMusicFiles(): FilesResponseDto

    @GET("api/v1/files/videos")
    suspend fun getVideoFiles(): FilesResponseDto

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

    @GET("api/v1/files/duplicate-files")
    suspend fun getDuplicateFiles(): List<FileItemDto>
}

@Serializable
data class FilesResponseDto(
    val files: List<FileItemDto> = emptyList(),
    val total: Int = 0,
)

@Serializable
data class FileItemDto(
    val id: String = "",
    val name: String = "",
    val path: String = "",
    val size: Long = 0,
    @SerialName("mime_type") val mimeType: String = "",
    @SerialName("is_dir") val isDir: Boolean = false,
    @SerialName("is_starred") val isStarred: Boolean = false,
    @SerialName("parent_id") val parentId: String? = null,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("modified_at") val modifiedAt: String = "",
)

@Serializable
data class CreateFolderRequest(val name: String, @SerialName("parent_id") val parentId: String? = null)

@Serializable
data class RenameRequest(val id: String, val name: String)

@Serializable
data class MoveRequest(val id: String, @SerialName("target_id") val targetId: String)
