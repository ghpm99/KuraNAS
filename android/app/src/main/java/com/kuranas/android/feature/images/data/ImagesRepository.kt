package com.kuranas.android.feature.images.data

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

class ImagesRepository @Inject constructor(
    private val api: ImageLibraryApi,
    private val serverStore: ServerStore,
) {
    suspend fun getImagesPage(cursor: String?): AppResult<ImageLibraryPageDto> = safeApiCall {
        api.getLibraryPage(cursor = cursor?.takeIf { it.isNotEmpty() })
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
    fun provideImageLibraryApi(retrofit: Retrofit): ImageLibraryApi = retrofit.create(ImageLibraryApi::class.java)

    @Provides
    @Singleton
    fun provideImagesRepository(api: ImageLibraryApi, serverStore: ServerStore) = ImagesRepository(api, serverStore)
}
