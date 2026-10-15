package com.kuranas.android.feature.video.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.paging.toListPage
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
    suspend fun getLibraryVideos(searchQuery: String, page: Int): AppResult<ListPage<VideoItemDto>> = safeApiCall {
        api.getLibraryVideos(page, LIBRARY_PAGE_SIZE, searchQuery.ifBlank { null }).toListPage()
    }
    suspend fun getPlaylists(): AppResult<List<VideoPlaylistDto>> = safeApiCall { api.getPlaylists() }
    suspend fun getPlaylistsBySection(
        section: VideoPlaylistSection,
        page: Int = 1,
        pageSize: Int = PLAYLIST_SECTION_PAGE_SIZE,
    ): AppResult<List<VideoPlaylistDto>> = safeApiCall { api.getPlaylistsBySection(section.path, page, pageSize).items }
    suspend fun getPlaylistHeader(id: Int): AppResult<VideoPlaylistDto> = safeApiCall { api.getPlaylistHeader(id) }
    suspend fun getPlaylistItems(id: Int, page: Int): AppResult<ListPage<VideoPlaylistItemDto>> = safeApiCall {
        api.getPlaylistItems(id, page, PLAYLIST_ITEMS_PAGE_SIZE).toListPage()
    }
    suspend fun startPlayback(videoId: String, playlistId: Int? = null): AppResult<PlaybackStateDto> = safeApiCall {
        api.startPlayback(StartPlaybackRequest(videoId.toIntOrNull() ?: 0, playlistId)).playbackState
    }
    suspend fun updatePlaybackState(
        videoId: String,
        playlistId: Int?,
        currentTimeSeconds: Double,
        durationSeconds: Double?,
        isPaused: Boolean,
    ): AppResult<Unit> = safeApiCall {
        api.updatePlaybackState(
            UpdatePlaybackRequest(
                playlistId = playlistId,
                videoId = videoId.toIntOrNull(),
                currentTime = currentTimeSeconds,
                duration = durationSeconds,
                isPaused = isPaused,
            ),
        )
    }
    suspend fun getPlaybackState(): AppResult<PlaybackStateDto> = safeApiCall { api.getPlaybackState().playbackState }
    suspend fun nextVideo(): AppResult<PlaybackStateDto> = safeApiCall { api.nextVideo().playbackState }
    suspend fun previousVideo(): AppResult<PlaybackStateDto> = safeApiCall { api.previousVideo().playbackState }

    suspend fun streamUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-stream/$videoId"
    }

    suspend fun thumbnailUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-thumbnail/$videoId"
    }
}

private const val PLAYLIST_SECTION_PAGE_SIZE = 24
private const val PLAYLIST_ITEMS_PAGE_SIZE = 50
private const val LIBRARY_PAGE_SIZE = 50

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object VideoModule {
    @Provides
    @Singleton
    fun provideVideoApi(retrofit: Retrofit): VideoApi = retrofit.create(VideoApi::class.java)
}
