package com.kuranas.android.feature.home.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

data class HomeStats(
    val totalSpaceUsed: Long,
    val totalFiles: Long,
    val totalDirectories: Long,
    val totalMusic: Long,
    val totalVideos: Long,
    val totalImages: Long,
)

class HomeRepository @Inject constructor(private val api: HomeApi) {

    suspend fun getStats(): AppResult<HomeStats> = safeApiCall {
        val space = api.getTotalSpaceUsed()
        val overview = api.getAnalyticsOverview()
        val dirs = api.getTotalDirectories()
        HomeStats(
            totalSpaceUsed = space.totalSpaceUsed,
            totalFiles = overview.totalFiles,
            totalDirectories = dirs.total,
            totalMusic = overview.totalMusic,
            totalVideos = overview.totalVideos,
            totalImages = overview.totalImages,
        )
    }

    suspend fun getRecentFiles(): AppResult<List<RecentFileDto>> = safeApiCall {
        api.getRecentFiles()
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object HomeModule {
    @Provides
    @Singleton
    fun provideHomeApi(retrofit: Retrofit): HomeApi = retrofit.create(HomeApi::class.java)
}
