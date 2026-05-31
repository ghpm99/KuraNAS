package com.kuranas.android.feature.video.data

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

class VideoRepository @Inject constructor(
    private val api: VideoApi,
    private val serverStore: ServerStore,
) {
    suspend fun getHomeCatalog(): AppResult<VideoHomeCatalogDto> = safeApiCall { api.getHomeCatalog() }
    suspend fun getLibraryVideos(): AppResult<List<VideoItemDto>> = safeApiCall { api.getLibraryVideos().items }
    suspend fun getPlaylists(): AppResult<List<VideoPlaylistDto>> = safeApiCall { api.getPlaylists() }
    suspend fun getPlaylistById(id: Int): AppResult<VideoPlaylistDto> = safeApiCall { api.getPlaylistById(id) }
    suspend fun startPlayback(videoId: String, playlistId: Int? = null): AppResult<PlaybackStateDto> = safeApiCall {
        api.startPlayback(StartPlaybackRequest(videoId.toIntOrNull() ?: 0, playlistId))
    }
    suspend fun nextVideo(): AppResult<PlaybackStateDto> = safeApiCall { api.nextVideo() }
    suspend fun previousVideo(): AppResult<PlaybackStateDto> = safeApiCall { api.previousVideo() }

    suspend fun streamUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-stream/$videoId"
    }

    suspend fun thumbnailUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-thumbnail/$videoId"
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object VideoModule {
    @Provides
    @Singleton
    fun provideVideoApi(retrofit: Retrofit): VideoApi = retrofit.create(VideoApi::class.java)
}
