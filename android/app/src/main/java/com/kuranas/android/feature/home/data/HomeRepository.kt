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

data class HomeData(
    val stats: HomeStats,
    val recentFiles: List<RecentFileDto>,
)

class HomeRepository @Inject constructor(private val api: HomeApi) {

    suspend fun getHomeData(): AppResult<HomeData> = safeApiCall {
        val space = api.getTotalSpaceUsed()
        val overview = api.getAnalyticsOverview()
        val dirs = api.getTotalDirectories()
        fun typeCount(vararg names: String): Long =
            overview.types.firstOrNull { it.type in names }?.count ?: 0
        val stats = HomeStats(
            totalSpaceUsed = space.totalSpaceUsed,
            totalFiles = overview.counts.filesTotal,
            totalDirectories = dirs.total,
            totalMusic = typeCount("audio", "music"),
            totalVideos = typeCount("video"),
            totalImages = typeCount("image", "images"),
        )
        HomeData(stats, overview.recentFiles)
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object HomeModule {
    @Provides
    @Singleton
    fun provideHomeApi(retrofit: Retrofit): HomeApi = retrofit.create(HomeApi::class.java)
}
