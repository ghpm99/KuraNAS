package com.kuranas.android.feature.search.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.Query

interface SearchApi {
    @GET("api/v1/search/global")
    suspend fun searchGlobal(@Query("q") query: String): SearchResultsDto
}

@Serializable
data class SearchResultsDto(
    val files: List<SearchFileDto> = emptyList(),
    val music: List<SearchFileDto> = emptyList(),
    val videos: List<SearchFileDto> = emptyList(),
    val total: Int = 0,
)

@Serializable
data class SearchFileDto(
    val id: String = "",
    val name: String = "",
    val path: String = "",
    val size: Long = 0,
    @SerialName("mime_type") val mimeType: String = "",
    @SerialName("is_dir") val isDir: Boolean = false,
)
