package com.kuranas.android.feature.search.data

import com.kuranas.android.core.network.mimeTypeForFormat
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.Query

interface SearchApi {
    @GET("api/v1/search/global")
    suspend fun searchGlobal(@Query("q") query: String): SearchResultsDto
}

/**
 * Espelha a resposta de /search/global. O backend devolve `files` misturando
 * áudio/vídeo/documentos; derivamos `music`/`videos`/`total` para a UI.
 */
@Serializable
data class SearchResultsDto(
    val query: String = "",
    @SerialName("files") val allFiles: List<SearchFileDto> = emptyList(),
    val folders: List<SearchFileDto> = emptyList(),
    @SerialName("videos") val videoMatches: List<SearchFileDto> = emptyList(),
    val images: List<SearchFileDto> = emptyList(),
    val artists: List<SearchGroupDto> = emptyList(),
    val albums: List<SearchGroupDto> = emptyList(),
    val playlists: List<SearchGroupDto> = emptyList(),
) {
    val files: List<SearchFileDto>
        get() = allFiles.filter {
            !it.mimeType.startsWith("audio/") && !it.mimeType.startsWith("video/")
        }
    val music: List<SearchFileDto> get() = allFiles.filter { it.mimeType.startsWith("audio/") }
    val videos: List<SearchFileDto> get() = videoMatches + allFiles.filter { it.mimeType.startsWith("video/") }
    val total: Int
        get() = allFiles.size + folders.size + videoMatches.size + images.size +
            artists.size + albums.size + playlists.size
}

@Serializable
data class SearchFileDto(
    @SerialName("id") val rawId: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val format: String = "",
    val starred: Boolean = false,
) {
    val id: String get() = rawId.toString()
    val isDir: Boolean get() = false
    val mimeType: String get() = mimeTypeForFormat(format)
}

/** Grupos de catálogo retornados na busca (artists/albums/playlists). */
@Serializable
data class SearchGroupDto(
    val key: String = "",
    val name: String = "",
    val artist: String = "",
    val album: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
)
