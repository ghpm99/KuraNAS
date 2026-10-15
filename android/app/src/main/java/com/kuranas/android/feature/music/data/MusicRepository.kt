package com.kuranas.android.feature.music.data

import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.PageDto
import com.kuranas.android.core.paging.toListPage
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
    suspend fun getAllTracks(page: Int): AppResult<ListPage<TrackDto>> = safeApiCall { api.getAllTracks(page).toListPage() }

    /**
     * Resolve uma faixa pelo id do arquivo consultando o servidor diretamente, para trazer
     * título/artista. Se a consulta falhar (sem conexão ou arquivo não indexado como música),
     * devolve uma faixa mínima: o player só precisa do id para montar a URL de stream.
     */
    suspend fun getTrackById(id: Int): TrackDto = when (val lookup = safeApiCall { api.getTrackById(id) }) {
        is AppResult.Success -> lookup.data
        is AppResult.Error -> TrackDto(id = id)
    }
    suspend fun getArtists(page: Int): AppResult<ListPage<ArtistDto>> = safeApiCall { api.getArtists(page).toListPage() }
    suspend fun getAlbums(page: Int): AppResult<ListPage<AlbumDto>> = safeApiCall { api.getAlbums(page).toListPage() }
    suspend fun getGenres(page: Int): AppResult<ListPage<GenreDto>> = safeApiCall { api.getGenres(page).toListPage() }
    suspend fun getTracksByArtist(key: String, page: Int): AppResult<ListPage<TrackDto>> = safeApiCall { api.getTracksByArtist(key, page).toListPage() }
    suspend fun getTracksByAlbum(key: String, page: Int): AppResult<ListPage<TrackDto>> = safeApiCall { api.getTracksByAlbum(key, page).toListPage() }
    suspend fun getTracksByGenre(key: String, page: Int): AppResult<ListPage<TrackDto>> = safeApiCall { api.getTracksByGenre(key, page).toListPage() }
    suspend fun getFolders(page: Int): AppResult<ListPage<FolderDto>> = safeApiCall { api.getFolders(page).toListPage() }
    suspend fun getTracksByFolder(key: String, page: Int): AppResult<ListPage<TrackDto>> = safeApiCall { api.getTracksByFolder(key, page).toListPage() }
    suspend fun getPlaylists(page: Int): AppResult<ListPage<PlaylistDto>> = safeApiCall { api.getPlaylists(page).toListPage() }
    suspend fun getPlaylistTracks(id: Int, page: Int): AppResult<ListPage<TrackDto>> = safeApiCall {
        val playlistPage = api.getPlaylistTracks(id, page)
        ListPage(playlistPage.items.map { it.file }, playlistPage.pagination.hasNext)
    }
    suspend fun createPlaylist(name: String): AppResult<PlaylistDto> = safeApiCall { api.createPlaylist(CreatePlaylistRequest(name)) }
    suspend fun deletePlaylist(id: Int): AppResult<Unit> = safeApiCall { api.deletePlaylist(id) }
    suspend fun addTrackToPlaylist(playlistId: Int, trackId: Int): AppResult<Unit> = safeApiCall { api.addTrackToPlaylist(playlistId, AddTrackRequest(trackId)) }
    suspend fun removeTrackFromPlaylist(playlistId: Int, trackId: Int): AppResult<Unit> = safeApiCall { api.removeTrackFromPlaylist(playlistId, trackId) }
    suspend fun getPlayerState(): AppResult<PlayerStateDto> = safeApiCall { api.getPlayerState() }

    suspend fun streamUrl(trackId: Int): String = "${baseUrl()}/api/v1/files/stream/$trackId"

    suspend fun thumbnailUrl(trackId: Int): String = "${baseUrl()}/api/v1/files/thumbnail/$trackId"

    /**
     * O valor salvo pelo usuário pode vir sem esquema/porta (ex.: "192.168.18.7:8000",
     * "192.168.18.7"). ExoPlayer/Coil precisam de uma URL absoluta válida, então
     * garantimos `http://` e a porta padrão 8000 quando ausentes — mesma lógica do
     * interceptor em NetworkModule.parseServerUrl.
     */
    private suspend fun baseUrl(): String {
        val raw = (serverStore.serverUrl.first() ?: "").trim().trimEnd('/')
        if (raw.isEmpty()) return ""
        val withScheme = if (raw.contains("://")) raw else "http://$raw"
        val authority = withScheme.substringAfter("://").substringBefore("/")
        val hasPort = authority.substringAfterLast(']').contains(":")
        return if (hasPort) withScheme else "$withScheme:8000"
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object MusicModule {
    @Provides
    @Singleton
    fun provideMusicApi(retrofit: Retrofit): MusicApi = retrofit.create(MusicApi::class.java)
}
