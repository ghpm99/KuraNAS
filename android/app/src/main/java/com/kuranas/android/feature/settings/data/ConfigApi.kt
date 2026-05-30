package com.kuranas.android.feature.settings.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.PUT

interface ConfigApi {
    @GET("api/v1/configuration/about")
    suspend fun getAbout(): AboutDto

    @GET("api/v1/configuration/settings")
    suspend fun getSettings(): ServerSettingsDto

    @PUT("api/v1/configuration/settings")
    suspend fun updateSettings(@Body body: ServerSettingsDto): ServerSettingsDto

    @GET("api/v1/update/status")
    suspend fun getUpdateStatus(): UpdateStatusDto
}

@Serializable
data class AboutDto(
    val version: String = "",
    val platform: String = "",
    @SerialName("go_version") val goVersion: String = "",
    val uptime: String = "",
)

@Serializable
data class ServerSettingsDto(
    val language: String = "pt",
    @SerialName("thumbnail_quality") val thumbnailQuality: Int = 80,
    @SerialName("scan_interval") val scanInterval: Int = 3600,
)

@Serializable
data class UpdateStatusDto(
    @SerialName("has_update") val hasUpdate: Boolean = false,
    @SerialName("current_version") val currentVersion: String = "",
    @SerialName("latest_version") val latestVersion: String? = null,
)
