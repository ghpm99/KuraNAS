package com.kuranas.android.feature.images.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.Query

interface ImageLibraryApi {
    @GET("api/v1/image/library")
    suspend fun getLibraryPage(
        @Query("cursor") cursor: String? = null,
        @Query("page_size") pageSize: Int = DEFAULT_PAGE_SIZE,
    ): ImageLibraryPageDto

    companion object {
        const val DEFAULT_PAGE_SIZE = 60
    }
}

@Serializable
data class ImageLibraryPageDto(
    val items: List<ImageLibraryItemDto> = emptyList(),
    @SerialName("next_cursor") val nextCursor: String = "",
    @SerialName("has_next") val hasNext: Boolean = false,
)

@Serializable
data class ImageLibraryItemDto(
    @SerialName("file_id") val fileId: Int = 0,
    val name: String = "",
) {
    val id: String get() = fileId.toString()
}
