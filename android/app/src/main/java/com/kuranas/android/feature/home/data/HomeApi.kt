package com.kuranas.android.feature.home.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET

interface HomeApi {
    @GET("api/v1/files/total-space-used")
    suspend fun getTotalSpaceUsed(): SpaceUsedDto

    @GET("api/v1/files/total-files")
    suspend fun getTotalFiles(): TotalDto

    @GET("api/v1/files/total-directory")
    suspend fun getTotalDirectories(): TotalDto

    @GET("api/v1/files/recent")
    suspend fun getRecentFiles(): List<RecentFileDto>

    @GET("api/v1/analytics/overview")
    suspend fun getAnalyticsOverview(): AnalyticsOverviewDto
}

@Serializable
data class SpaceUsedDto(@SerialName("total_space_used") val totalSpaceUsed: Long = 0)

@Serializable
data class TotalDto(val total: Long = 0)

@Serializable
data class RecentFileDto(
    val id: String = "",
    val name: String = "",
    val path: String = "",
    val size: Long = 0,
    @SerialName("mime_type") val mimeType: String = "",
    @SerialName("accessed_at") val accessedAt: String = "",
)

@Serializable
data class AnalyticsOverviewDto(
    @SerialName("total_files") val totalFiles: Long = 0,
    @SerialName("total_size") val totalSize: Long = 0,
    @SerialName("total_music") val totalMusic: Long = 0,
    @SerialName("total_videos") val totalVideos: Long = 0,
    @SerialName("total_images") val totalImages: Long = 0,
)
