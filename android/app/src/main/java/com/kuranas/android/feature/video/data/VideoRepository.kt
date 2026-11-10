package com.kuranas.android.feature.video.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.paging.toListPage
import com.kuranas.android.core.server.ServerStore
import com.kuranas.android.feature.files.data.FileStarStore
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
    private val starStore: FileStarStore,
) {
    suspend fun getHomeCatalog(): AppResult<VideoHomeCatalogDto> = safeApiCall { api.getHomeCatalog() }.recordingCatalogStars()
    suspend fun getLibraryVideos(searchQuery: String, page: Int): AppResult<ListPage<VideoItemDto>> = safeApiCall {
        api.getLibraryVideos(page, LIBRARY_PAGE_SIZE, searchQuery.ifBlank { null }).toListPage()
    }.recordingVideoStars()
    suspend fun getPlaylists(): AppResult<List<VideoPlaylistDto>> = safeApiCall { api.getPlaylists() }
    suspend fun getPlaylistsBySection(
        section: VideoPlaylistSection,
        page: Int = 1,
        pageSize: Int = PLAYLIST_SECTION_PAGE_SIZE,
    ): AppResult<List<VideoPlaylistDto>> = safeApiCall { api.getPlaylistsBySection(section.path, page, pageSize).items }
    suspend fun getPlaylistHeader(id: Int): AppResult<VideoPlaylistDto> = safeApiCall { api.getPlaylistHeader(id) }
    suspend fun getPlaylistItems(id: Int, page: Int): AppResult<ListPage<VideoPlaylistItemDto>> = safeApiCall {
        api.getPlaylistItems(id, page, PLAYLIST_ITEMS_PAGE_SIZE).toListPage()
    }.recordingPlaylistItemStars()
    suspend fun startPlayback(videoId: String, playlistId: Int? = null): AppResult<PlaybackSessionDto> = safeApiCall {
        api.startPlayback(StartPlaybackRequest(videoId.toIntOrNull() ?: 0, playlistId))
    }.recordingSessionStars()
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
    suspend fun nextVideo(): AppResult<PlaybackSessionDto> = safeApiCall { api.nextVideo() }.recordingSessionStars()
    suspend fun previousVideo(): AppResult<PlaybackSessionDto> = safeApiCall { api.previousVideo() }.recordingSessionStars()

    suspend fun streamUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-stream/$videoId"
    }

    suspend fun thumbnailUrl(videoId: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/video-thumbnail/$videoId"
    }

    private fun AppResult<ListPage<VideoItemDto>>.recordingVideoStars(): AppResult<ListPage<VideoItemDto>> {
        if (this is AppResult.Success) recordStars(data.items)
        return this
    }

    private fun AppResult<ListPage<VideoPlaylistItemDto>>.recordingPlaylistItemStars(): AppResult<ListPage<VideoPlaylistItemDto>> {
        if (this is AppResult.Success) recordStars(data.items.map { playlistItem -> playlistItem.video })
        return this
    }

    private fun AppResult<VideoHomeCatalogDto>.recordingCatalogStars(): AppResult<VideoHomeCatalogDto> {
        if (this is AppResult.Success) {
            recordStars(data.sections.flatMap { section -> section.items.map { catalogItem -> catalogItem.video } })
        }
        return this
    }

    private fun AppResult<PlaybackSessionDto>.recordingSessionStars(): AppResult<PlaybackSessionDto> {
        if (this is AppResult.Success) recordStars(data.playlist.items.map { playlistItem -> playlistItem.video })
        return this
    }

    private fun recordStars(videos: List<VideoItemDto>) {
        starStore.record(videos.associate { video -> video.id to video.starred })
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
