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
    @SerialName("commit_hash") val commitHash: String = "",
    val platform: String = "",
    val path: String = "",
    val lang: String = "",
    @SerialName("enable_workers") val enableWorkers: Boolean = false,
    // O backend expõe o horário de inicialização na chave (com typo) `statup_time`.
    @SerialName("statup_time") val startupTime: String = "",
    @SerialName("go_version") val goVersion: String = "",
    @SerialName("node_version") val nodeVersion: String = "",
) {
    /** Mantém a API esperada pela UI; o backend só informa o horário de início. */
    val uptime: String get() = startupTime
}

/**
 * Espelha configuration.SettingsDto: estrutura aninhada
 * { library, indexing, players, appearance, language:{current, available} }.
 */
@Serializable
data class ServerSettingsDto(
    val library: LibrarySettingsDto = LibrarySettingsDto(),
    val indexing: IndexingSettingsDto = IndexingSettingsDto(),
    val players: PlayerSettingsDto = PlayerSettingsDto(),
    val appearance: AppearanceSettingsDto = AppearanceSettingsDto(),
    val language: LanguageSettingsDto = LanguageSettingsDto(),
)

@Serializable
data class LibrarySettingsDto(
    @SerialName("runtime_root_path") val runtimeRootPath: String = "",
    @SerialName("watched_paths") val watchedPaths: List<String> = emptyList(),
    @SerialName("remember_last_location") val rememberLastLocation: Boolean = false,
    @SerialName("prioritize_favorites") val prioritizeFavorites: Boolean = false,
)

@Serializable
data class IndexingSettingsDto(
    @SerialName("workers_enabled") val workersEnabled: Boolean = false,
    @SerialName("scan_on_startup") val scanOnStartup: Boolean = false,
    @SerialName("extract_metadata") val extractMetadata: Boolean = false,
    @SerialName("generate_previews") val generatePreviews: Boolean = false,
)

@Serializable
data class PlayerSettingsDto(
    @SerialName("remember_music_queue") val rememberMusicQueue: Boolean = false,
    @SerialName("remember_video_progress") val rememberVideoProgress: Boolean = false,
    @SerialName("autoplay_next_video") val autoplayNextVideo: Boolean = false,
    @SerialName("image_slideshow_seconds") val imageSlideshowSeconds: Int = 4,
)

@Serializable
data class AppearanceSettingsDto(
    @SerialName("accent_color") val accentColor: String = "violet",
    @SerialName("reduce_motion") val reduceMotion: Boolean = false,
)

@Serializable
data class LanguageSettingsDto(
    val current: String = "",
    val available: List<String> = emptyList(),
)

@Serializable
data class UpdateStatusDto(
    @SerialName("has_update") val hasUpdate: Boolean = false,
    @SerialName("current_version") val currentVersion: String = "",
    @SerialName("latest_version") val latestVersion: String? = null,
)
