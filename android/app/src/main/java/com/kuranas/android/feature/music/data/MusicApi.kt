package com.kuranas.android.feature.music.data

import com.kuranas.android.core.network.PageDto
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
    suspend fun getAllTracks(): PageDto<TrackDto>

    @GET("api/v1/music/library/home")
    suspend fun getHomeCatalog(): MusicHomeCatalogDto

    @GET("api/v1/music/library/artists")
    suspend fun getArtists(): PageDto<ArtistDto>

    @GET("api/v1/music/library/artists/{key}/tracks")
    suspend fun getTracksByArtist(@Path("key") key: String): PageDto<TrackDto>

    @GET("api/v1/music/library/albums")
    suspend fun getAlbums(): PageDto<AlbumDto>

    @GET("api/v1/music/library/albums/{key}/tracks")
    suspend fun getTracksByAlbum(@Path("key") key: String): PageDto<TrackDto>

    @GET("api/v1/music/library/genres")
    suspend fun getGenres(): PageDto<GenreDto>

    @GET("api/v1/music/library/genres/{key}/tracks")
    suspend fun getTracksByGenre(@Path("key") key: String): PageDto<TrackDto>

    @GET("api/v1/music/library/folders")
    suspend fun getFolders(): PageDto<FolderDto>

    // Playlists
    @GET("api/v1/music/playlists/")
    suspend fun getPlaylists(): PageDto<PlaylistDto>

    @POST("api/v1/music/playlists/")
    suspend fun createPlaylist(@Body body: CreatePlaylistRequest): PlaylistDto

    // O backend retorna um array puro aqui (não o envelope PageDto).
    @GET("api/v1/music/playlists/system")
    suspend fun getSystemPlaylists(): List<PlaylistDto>

    @GET("api/v1/music/playlists/{id}")
    suspend fun getPlaylistById(@Path("id") id: Int): PlaylistDto

    @DELETE("api/v1/music/playlists/{id}")
    suspend fun deletePlaylist(@Path("id") id: Int)

    @GET("api/v1/music/playlists/{id}/tracks")
    suspend fun getPlaylistTracks(@Path("id") id: Int): PageDto<PlaylistTrackDto>

    @POST("api/v1/music/playlists/{id}/tracks")
    suspend fun addTrackToPlaylist(@Path("id") id: Int, @Body body: AddTrackRequest)

    @DELETE("api/v1/music/playlists/{id}/tracks/{fileId}")
    suspend fun removeTrackFromPlaylist(@Path("id") id: Int, @Path("fileId") fileId: Int)

    // Player state
    @GET("api/v1/music/player-state/")
    suspend fun getPlayerState(): PlayerStateDto

    @PUT("api/v1/music/player-state/")
    suspend fun updatePlayerState(@Body body: UpdatePlayerStateRequest)
}

/** Uma faixa é um arquivo (files.FileDto) com metadados de música em [metadata]. */
@Serializable
data class TrackDto(
    val id: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val format: String = "",
    val size: Long = 0,
    val metadata: TrackMetadataDto? = null,
) {
    val title: String get() = metadata?.title?.takeIf { it.isNotBlank() } ?: name
    val artist: String? get() = metadata?.artist?.takeIf { it.isNotBlank() }
    val album: String? get() = metadata?.album?.takeIf { it.isNotBlank() }
    val durationSeconds: Int? get() = metadata?.length?.toInt()?.takeIf { it > 0 }
}

/**
 * Metadata de áudio real do backend: `year`/`track_number` vêm como STRING (podem
 * ser vazias) e a duração é `length` (float, em segundos) — não inteiros.
 */
@Serializable
data class TrackMetadataDto(
    val title: String = "",
    val artist: String = "",
    val album: String = "",
    val genre: String = "",
    val year: String = "",
    @SerialName("track_number") val trackNumber: String = "",
    val length: Double = 0.0,
    val mime: String = "",
)

@Serializable
data class ArtistDto(
    val key: String = "",
    val artist: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
    @SerialName("album_count") val albumCount: Int = 0,
)

@Serializable
data class AlbumDto(
    val key: String = "",
    val album: String = "",
    val artist: String = "",
    val year: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
)

@Serializable
data class GenreDto(
    val key: String = "",
    val genre: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
)

@Serializable
data class FolderDto(
    val folder: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
)

@Serializable
data class PlaylistDto(
    val id: Int = 0,
    val name: String = "",
    val description: String = "",
    @SerialName("is_system") val isSystem: Boolean = false,
    @SerialName("is_auto") val isAuto: Boolean = false,
    val kind: String = "",
    @SerialName("source_key") val sourceKey: String = "",
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    @SerialName("track_count") val trackCount: Int = 0,
)

@Serializable
data class PlaylistTrackDto(
    val id: Int = 0,
    val position: Int = 0,
    @SerialName("added_at") val addedAt: String = "",
    val file: TrackDto = TrackDto(),
)

@Serializable
data class MusicLibrarySummaryDto(
    @SerialName("total_tracks") val totalTracks: Int = 0,
    @SerialName("total_artists") val totalArtists: Int = 0,
    @SerialName("total_albums") val totalAlbums: Int = 0,
    @SerialName("total_genres") val totalGenres: Int = 0,
    @SerialName("total_folders") val totalFolders: Int = 0,
)

@Serializable
data class MusicHomeCatalogDto(
    val summary: MusicLibrarySummaryDto = MusicLibrarySummaryDto(),
    val playlists: List<PlaylistDto> = emptyList(),
    val artists: List<ArtistDto> = emptyList(),
    val albums: List<AlbumDto> = emptyList(),
)

@Serializable
data class PlayerStateDto(
    val id: Int = 0,
    @SerialName("client_id") val clientId: String = "",
    @SerialName("playlist_id") val playlistId: Int? = null,
    @SerialName("current_file_id") val currentFileId: Int? = null,
    @SerialName("current_position") val currentPosition: Double = 0.0,
    val volume: Double = 1.0,
    val shuffle: Boolean = false,
    @SerialName("repeat_mode") val repeatMode: String = "none",
    @SerialName("updated_at") val updatedAt: String = "",
)

@Serializable
data class UpdatePlayerStateRequest(
    @SerialName("playlist_id") val playlistId: Int? = null,
    @SerialName("current_file_id") val currentFileId: Int? = null,
    @SerialName("current_position") val currentPosition: Double = 0.0,
    val volume: Double = 1.0,
    val shuffle: Boolean = false,
    @SerialName("repeat_mode") val repeatMode: String = "none",
)

@Serializable
data class CreatePlaylistRequest(val name: String, val description: String = "")

@Serializable
data class AddTrackRequest(@SerialName("file_id") val fileId: Int)
