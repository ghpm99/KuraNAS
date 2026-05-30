package com.kuranas.android.feature.images.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.server.ServerStore
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesApi
import dagger.Module
import dagger.Provides
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.first
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class ImagesRepository @Inject constructor(
    private val api: FilesApi,
    private val serverStore: ServerStore,
) {
    suspend fun getImages(): AppResult<List<FileItemDto>> = safeApiCall {
        api.getImages().files
    }

    suspend fun getThumbnailUrl(id: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/thumbnail/$id"
    }

    suspend fun getBlobUrl(id: String): String {
        val base = serverStore.serverUrl.first() ?: ""
        return "$base/api/v1/files/blob/$id"
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object ImagesModule {
    @Provides
    @Singleton
    fun provideImagesRepository(api: FilesApi, serverStore: ServerStore) = ImagesRepository(api, serverStore)
}
