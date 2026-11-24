package com.kuranas.android.feature.documents.data

import com.kuranas.android.core.network.PageDto
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.Query

interface DocumentsApi {
    @GET("api/v1/documents")
    suspend fun getDocuments(
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
        @Query("format") format: String? = null,
        @Query("folder_id") folderId: Int? = null,
    ): PageDto<DocumentDto>

    @GET("api/v1/documents/count")
    suspend fun getDocumentsCount(
        @Query("format") format: String? = null,
        @Query("folder_id") folderId: Int? = null,
    ): DocumentsCountDto

    @GET("api/v1/documents/search")
    suspend fun searchDocuments(
        @Query("q") query: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
        @Query("format") format: String? = null,
    ): PageDto<DocumentSearchResultDto>
}

@Serializable
data class DocumentDto(
    val id: Int = 0,
    val name: String = "",
    val format: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val size: Long = 0,
    @SerialName("modified_at") val modifiedAt: String = "",
    @SerialName("created_at") val createdAt: String = "",
    val starred: Boolean = false,
    @SerialName("index_state") val indexState: String = "",
)

@Serializable
data class DocumentsCountDto(val total: Int = 0)

@Serializable
data class DocumentSearchResultDto(
    @SerialName("file_id") val fileId: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val format: String = "",
    val size: Long = 0,
    @SerialName("updated_at") val updatedAt: String = "",
    val starred: Boolean = false,
    val snippet: String = "",
    val page: Int = 0,
)
