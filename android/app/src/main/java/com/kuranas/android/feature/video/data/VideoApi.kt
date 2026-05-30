package com.kuranas.android.feature.video.data

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
    suspend fun getLibraryVideos(): VideoLibraryDto

    @GET("api/v1/video/playlists/")
    suspend fun getPlaylists(): List<VideoPlaylistDto>

    @GET("api/v1/video/playlists/{id}")
    suspend fun getPlaylistById(@Path("id") id: Int): VideoPlaylistDto

    @GET("api/v1/video/playlists/{id}")
    suspend fun getPlaylistVideos(@Path("id") id: Int): VideoLibraryDto

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
    suspend fun updatePlaybackState(@Body body: PlaybackStateDto)

    @POST("api/v1/video/playback/next")
    suspend fun nextVideo(): PlaybackStateDto

    @POST("api/v1/video/playback/previous")
    suspend fun previousVideo(): PlaybackStateDto
}

@Serializable
data class VideoLibraryDto(val files: List<VideoItemDto> = emptyList(), val total: Int = 0)

@Serializable
data class VideoItemDto(
    val id: String = "",
    val name: String = "",
    val path: String = "",
    val size: Long = 0,
    @SerialName("mime_type") val mimeType: String = "",
    val duration: Int? = null,
    @SerialName("thumbnail_path") val thumbnailPath: String? = null,
)

@Serializable
data class VideoPlaylistDto(
    val id: Int = 0,
    val name: String = "",
    val hidden: Boolean = false,
    val count: Int = 0,
    val videos: List<VideoItemDto> = emptyList(),
)

@Serializable
data class VideoHomeCatalogDto(
    val playlists: List<VideoPlaylistDto> = emptyList(),
    @SerialName("recent_videos") val recentVideos: List<VideoItemDto> = emptyList(),
)

@Serializable
data class PlaybackStateDto(
    @SerialName("video_id") val videoId: String? = null,
    @SerialName("playlist_id") val playlistId: Int? = null,
    val position: Long = 0,
    @SerialName("is_playing") val isPlaying: Boolean = false,
)

@Serializable
data class StartPlaybackRequest(
    @SerialName("video_id") val videoId: String,
    @SerialName("playlist_id") val playlistId: Int? = null,
)

@Serializable
data class UpdatePlaylistRequest(val name: String)

@Serializable
data class HiddenRequest(val hidden: Boolean)

@Serializable
data class AddVideoRequest(@SerialName("video_id") val videoId: String)
