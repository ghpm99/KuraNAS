package com.kuranas.android.feature.music.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

interface MusicApi {
    // Library
    @GET("api/v1/music/library/")
    suspend fun getAllTracks(): LibraryTracksDto

    @GET("api/v1/music/library/home")
    suspend fun getHomeCatalog(): MusicHomeCatalogDto

    @GET("api/v1/music/library/artists")
    suspend fun getArtists(): List<ArtistDto>

    @GET("api/v1/music/library/artists/{key}/tracks")
    suspend fun getTracksByArtist(@Path("key") key: String): LibraryTracksDto

    @GET("api/v1/music/library/albums")
    suspend fun getAlbums(): List<AlbumDto>

    @GET("api/v1/music/library/albums/{key}/tracks")
    suspend fun getTracksByAlbum(@Path("key") key: String): LibraryTracksDto

    @GET("api/v1/music/library/genres")
    suspend fun getGenres(): List<GenreDto>

    @GET("api/v1/music/library/genres/{key}/tracks")
    suspend fun getTracksByGenre(@Path("key") key: String): LibraryTracksDto

    @GET("api/v1/music/library/folders")
    suspend fun getFolders(): List<FolderDto>

    // Playlists
    @GET("api/v1/music/playlists/")
    suspend fun getPlaylists(): List<PlaylistDto>

    @POST("api/v1/music/playlists/")
    suspend fun createPlaylist(@Body body: CreatePlaylistRequest): PlaylistDto

    @GET("api/v1/music/playlists/now-playing")
    suspend fun getNowPlaying(): NowPlayingDto

    @GET("api/v1/music/playlists/system")
    suspend fun getSystemPlaylists(): List<PlaylistDto>

    @GET("api/v1/music/playlists/{id}")
    suspend fun getPlaylistById(@Path("id") id: Int): PlaylistDto

    @DELETE("api/v1/music/playlists/{id}")
    suspend fun deletePlaylist(@Path("id") id: Int)

    @GET("api/v1/music/playlists/{id}/tracks")
    suspend fun getPlaylistTracks(@Path("id") id: Int): LibraryTracksDto

    @POST("api/v1/music/playlists/{id}/tracks")
    suspend fun addTrackToPlaylist(@Path("id") id: Int, @Body body: AddTrackRequest)

    @DELETE("api/v1/music/playlists/{id}/tracks/{fileId}")
    suspend fun removeTrackFromPlaylist(@Path("id") id: Int, @Path("fileId") fileId: String)

    // Player state
    @GET("api/v1/music/player-state/")
    suspend fun getPlayerState(): PlayerStateDto

    @PUT("api/v1/music/player-state/")
    suspend fun updatePlayerState(@Body body: PlayerStateDto)
}

@Serializable
data class LibraryTracksDto(val tracks: List<TrackDto> = emptyList(), val total: Int = 0)

@Serializable
data class TrackDto(
    val id: String = "",
    val name: String = "",
    val path: String = "",
    val size: Long = 0,
    @SerialName("mime_type") val mimeType: String = "",
    val artist: String? = null,
    val album: String? = null,
    val genre: String? = null,
    val duration: Int? = null,
    @SerialName("track_number") val trackNumber: Int? = null,
    val year: Int? = null,
)

@Serializable
data class ArtistDto(val name: String = "", val count: Int = 0)

@Serializable
data class AlbumDto(val name: String = "", val artist: String? = null, val count: Int = 0)

@Serializable
data class GenreDto(val name: String = "", val count: Int = 0)

@Serializable
data class FolderDto(val name: String = "", val path: String = "", val count: Int = 0)

@Serializable
data class PlaylistDto(
    val id: Int = 0,
    val name: String = "",
    val count: Int = 0,
    @SerialName("created_at") val createdAt: String = "",
)

@Serializable
data class NowPlayingDto(
    @SerialName("track_id") val trackId: String? = null,
    @SerialName("playlist_id") val playlistId: Int? = null,
)

@Serializable
data class MusicHomeCatalogDto(
    @SerialName("recent_artists") val recentArtists: List<ArtistDto> = emptyList(),
    @SerialName("recent_albums") val recentAlbums: List<AlbumDto> = emptyList(),
)

@Serializable
data class PlayerStateDto(
    @SerialName("track_id") val trackId: String? = null,
    @SerialName("playlist_id") val playlistId: Int? = null,
    val position: Long = 0,
    @SerialName("is_playing") val isPlaying: Boolean = false,
    val shuffle: Boolean = false,
    val repeat: String = "none",
)

@Serializable
data class CreatePlaylistRequest(val name: String)

@Serializable
data class AddTrackRequest(@SerialName("file_id") val fileId: String)
