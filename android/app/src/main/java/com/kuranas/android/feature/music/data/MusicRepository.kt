package com.kuranas.android.feature.music.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.server.ServerStore
import dagger.Module
import dagger.Provides
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.first
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class MusicRepository @Inject constructor(
    private val api: MusicApi,
    private val serverStore: ServerStore,
) {
    suspend fun getAllTracks(): AppResult<List<TrackDto>> = safeApiCall { api.getAllTracks().items }
    suspend fun getArtists(): AppResult<List<ArtistDto>> = safeApiCall { api.getArtists().items }
    suspend fun getAlbums(): AppResult<List<AlbumDto>> = safeApiCall { api.getAlbums().items }
    suspend fun getGenres(): AppResult<List<GenreDto>> = safeApiCall { api.getGenres().items }
    suspend fun getTracksByArtist(key: String): AppResult<List<TrackDto>> = safeApiCall { api.getTracksByArtist(key).items }
    suspend fun getTracksByAlbum(key: String): AppResult<List<TrackDto>> = safeApiCall { api.getTracksByAlbum(key).items }
    suspend fun getTracksByGenre(key: String): AppResult<List<TrackDto>> = safeApiCall { api.getTracksByGenre(key).items }
    suspend fun getPlaylists(): AppResult<List<PlaylistDto>> = safeApiCall { api.getPlaylists().items }
    suspend fun getPlaylistTracks(id: Int): AppResult<List<TrackDto>> = safeApiCall { api.getPlaylistTracks(id).items.map { it.file } }
    suspend fun createPlaylist(name: String): AppResult<PlaylistDto> = safeApiCall { api.createPlaylist(CreatePlaylistRequest(name)) }
    suspend fun deletePlaylist(id: Int): AppResult<Unit> = safeApiCall { api.deletePlaylist(id) }
    suspend fun addTrackToPlaylist(playlistId: Int, trackId: Int): AppResult<Unit> = safeApiCall { api.addTrackToPlaylist(playlistId, AddTrackRequest(trackId)) }
    suspend fun removeTrackFromPlaylist(playlistId: Int, trackId: Int): AppResult<Unit> = safeApiCall { api.removeTrackFromPlaylist(playlistId, trackId) }
    suspend fun getPlayerState(): AppResult<PlayerStateDto> = safeApiCall { api.getPlayerState() }

    suspend fun streamUrl(trackId: Int): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/stream/$trackId"
    }

    suspend fun thumbnailUrl(trackId: Int): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/thumbnail/$trackId"
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object MusicModule {
    @Provides
    @Singleton
    fun provideMusicApi(retrofit: Retrofit): MusicApi = retrofit.create(MusicApi::class.java)
}
