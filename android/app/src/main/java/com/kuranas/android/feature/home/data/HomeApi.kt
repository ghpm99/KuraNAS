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
    suspend fun getTotalDirectories(): TotalDirectoryDto

    @GET("api/v1/analytics/overview")
    suspend fun getAnalyticsOverview(): AnalyticsOverviewDto
}

@Serializable
data class SpaceUsedDto(@SerialName("total_space_used") val totalSpaceUsed: Long = 0)

@Serializable
data class TotalDto(val total: Long = 0)

@Serializable
data class TotalDirectoryDto(@SerialName("total_directory") val total: Long = 0)

/** Espelha analytics.RecentFileDto (campo `recent_files` do /analytics/overview). */
@Serializable
data class RecentFileDto(
    val id: Int = 0,
    val name: String = "",
    val path: String = "",
    @SerialName("parent_path") val parentPath: String = "",
    val format: String = "",
    @SerialName("size_bytes") val sizeBytes: Long = 0,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
) {
    val size: Long get() = sizeBytes
    val mimeType: String get() = format
}

@Serializable
data class AnalyticsOverviewDto(
    val storage: AnalyticsStorageDto = AnalyticsStorageDto(),
    val counts: AnalyticsCountsDto = AnalyticsCountsDto(),
    val types: List<AnalyticsTypeDto> = emptyList(),
    @SerialName("recent_files") val recentFiles: List<RecentFileDto> = emptyList(),
)

@Serializable
data class AnalyticsStorageDto(
    @SerialName("total_bytes") val totalBytes: Long = 0,
    @SerialName("used_bytes") val usedBytes: Long = 0,
    @SerialName("free_bytes") val freeBytes: Long = 0,
    @SerialName("growth_bytes") val growthBytes: Long = 0,
)

@Serializable
data class AnalyticsCountsDto(
    @SerialName("files_total") val filesTotal: Long = 0,
    @SerialName("files_added") val filesAdded: Long = 0,
    val folders: Long = 0,
)

@Serializable
data class AnalyticsTypeDto(
    val type: String = "",
    val count: Long = 0,
    val bytes: Long = 0,
)
