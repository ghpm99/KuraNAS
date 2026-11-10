package com.kuranas.android.feature.trash.data

import com.kuranas.android.core.network.PageDto
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

interface TrashApi {
    @GET("api/v1/trash")
    suspend fun getTrashItems(
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): PageDto<TrashItemDto>

    @POST("api/v1/trash/{id}/restore")
    suspend fun restoreItem(@Path("id") id: Int)

    @DELETE("api/v1/trash/{id}")
    suspend fun deleteItemForever(@Path("id") id: Int)

    @DELETE("api/v1/trash")
    suspend fun emptyTrash()
}

/** Mirrors trash.TrashItemDto from the Go backend. */
@Serializable
data class TrashItemDto(
    val id: Int = 0,
    val name: String = "",
    @SerialName("original_path") val originalPath: String = "",
    val size: Long = 0,
    @SerialName("deleted_at") val deletedAt: String = "",
    @SerialName("expires_at") val expiresAt: String = "",
    @SerialName("is_directory") val isDirectory: Boolean = false,
)
