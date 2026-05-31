package com.kuranas.android.feature.video.data

import com.kuranas.android.core.network.PageDto
import com.kuranas.android.core.network.mimeTypeForFormat
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

interface VideoApi {
    @GET("api/v1/video/catalog/home")
    suspend fun getHomeCatalog(): VideoHomeCatalogDto

    @GET("api/v1/video/library/files")
    suspend fun getLibraryVideos(): PageDto<VideoItemDto>

    @GET("api/v1/video/playlists/")
    suspend fun getPlaylists(): List<VideoPlaylistDto>

    @GET("api/v1/video/playlists/{id}")
    suspend fun getPlaylistById(@Path("id") id: Int): VideoPlaylistDto

    @PUT("api/v1/video/playlists/{id}")
    suspend fun updatePlaylist(@Path("id") id: Int, @Body body: UpdatePlaylistRequest): VideoPlaylistDto

    @PUT("api/v1/video/playlists/{id}/hidden")
    suspend fun setPlaylistHidden(@Path("id") id: Int, @Body body: HiddenRequest)

    @POST("api/v1/video/playlists/{id}/videos")
    suspend fun addVideoToPlaylist(@Path("id") id: Int, @Body body: AddVideoRequest)

    @DELETE("api/v1/video/playlists/{id}/videos/{videoId}")
    suspend fun removeVideoFromPlaylist(@Path("id") id: Int, @Path("videoId") videoId: String)

    @POST("api/v1/video/playback/start")
    suspend fun startPlayback(@Body body: StartPlaybackRequest): PlaybackStateDto

    @GET("api/v1/video/playback/state")
    suspend fun getPlaybackState(): PlaybackStateDto

    @PUT("api/v1/video/playback/state")
    suspend fun updatePlaybackState(@Body body: UpdatePlaybackRequest)

    @POST("api/v1/video/playback/next")
    suspend fun nextVideo(): PlaybackStateDto

    @POST("api/v1/video/playback/previous")
    suspend fun previousVideo(): PlaybackStateDto
}

/** Espelha video.VideoFileDto. Props computadas preservam a API da UI. */
@Serializable
data class VideoItemDto(
    @SerialName("id") val rawId: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val format: String = "",
    val size: Long = 0,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
) {
    val id: String get() = rawId.toString()
    val mimeType: String get() = mimeTypeForFormat(format)
}

/** Espelha video.VideoPlaylistDto (com items[].video). */
@Serializable
data class VideoPlaylistDto(
    val id: Int = 0,
    val name: String = "",
    @SerialName("is_hidden") val isHidden: Boolean = false,
    @SerialName("item_count") val itemCount: Int = 0,
    val items: List<VideoPlaylistItemDto> = emptyList(),
) {
    val hidden: Boolean get() = isHidden
    val count: Int get() = itemCount
    val videos: List<VideoItemDto> get() = items.map { it.video }
}

@Serializable
data class VideoPlaylistItemDto(
    val video: VideoItemDto = VideoItemDto(),
    val status: String = "",
    @SerialName("progress_pct") val progressPct: Double = 0.0,
)

/**
 * Espelha video.VideoCatalogHomeDto (`{ sections: [{ key, title, items: [{ video }] }] }`).
 * `recentVideos` achata os vídeos de todas as seções para a UI.
 */
@Serializable
data class VideoHomeCatalogDto(
    val sections: List<VideoSectionDto> = emptyList(),
) {
    val recentVideos: List<VideoItemDto> get() = sections.flatMap { section -> section.items.map { it.video } }
    val playlists: List<VideoPlaylistDto> get() = emptyList()
}

@Serializable
data class VideoSectionDto(
    val key: String = "",
    val title: String = "",
    val items: List<VideoCatalogItemDto> = emptyList(),
)

@Serializable
data class VideoCatalogItemDto(
    val video: VideoItemDto = VideoItemDto(),
    val status: String = "",
    @SerialName("progress_pct") val progressPct: Double = 0.0,
)

/** Espelha video.VideoPlaybackStateDto. */
@Serializable
data class PlaybackStateDto(
    val id: Int = 0,
    @SerialName("client_id") val clientId: String = "",
    @SerialName("playlist_id") val playlistId: Int? = null,
    @SerialName("video_id") val videoId: Int? = null,
    @SerialName("current_time") val currentTime: Double = 0.0,
    val duration: Double = 0.0,
    @SerialName("is_paused") val isPaused: Boolean = false,
    val completed: Boolean = false,
)

@Serializable
data class StartPlaybackRequest(
    @SerialName("video_id") val videoId: Int,
    @SerialName("playlist_id") val playlistId: Int? = null,
)

@Serializable
data class UpdatePlaybackRequest(
    @SerialName("playlist_id") val playlistId: Int? = null,
    @SerialName("video_id") val videoId: Int? = null,
    @SerialName("current_time") val currentTime: Double? = null,
    val duration: Double? = null,
    @SerialName("is_paused") val isPaused: Boolean? = null,
    val completed: Boolean? = null,
)

@Serializable
data class UpdatePlaylistRequest(val name: String)

@Serializable
data class HiddenRequest(val hidden: Boolean)

@Serializable
data class AddVideoRequest(@SerialName("video_id") val videoId: Int)
