package com.kuranas.android.feature.ingest.data

import com.kuranas.android.core.network.PageDto
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

interface IngestApi {
    @GET("api/v1/ingest/targets")
    suspend fun getTargets(): List<IngestTargetDto>

    @GET("api/v1/ingest/presets")
    suspend fun getPresets(): List<IngestPresetDto>

    @POST("api/v1/ingest/fetch")
    suspend fun fetch(@Body request: IngestFetchRequestDto): IngestFetchResponseDto

    @GET("api/v1/ingest/downloads")
    suspend fun getDownloads(
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): PageDto<IngestDownloadDto>

    @POST("api/v1/ingest/downloads/{id}/retry")
    suspend fun retryDownload(@Path("id") downloadId: Int)

    @POST("api/v1/jobs/{id}/cancel")
    suspend fun cancelJob(@Path("id") jobId: Int)
}

@Serializable
data class IngestTargetDto(
    val label: String = "",
    val path: String = "",
)

@Serializable
data class IngestPresetDto(
    val key: String = "",
    val label: String = "",
)

@Serializable
data class IngestFetchRequestDto(
    val url: String,
    val preset: String,
    @SerialName("target_root") val targetRoot: String,
    val subfolder: String,
)

@Serializable
data class IngestFetchResponseDto(
    @SerialName("job_id") val jobId: Int = 0,
    @SerialName("download_id") val downloadId: Int = 0,
)

@Serializable
data class IngestDownloadDto(
    val id: Int = 0,
    @SerialName("job_id") val jobId: Int? = null,
    val url: String = "",
    val title: String? = null,
    val preset: String = "",
    @SerialName("target_root_id") val targetRootId: Int = 0,
    @SerialName("target_root_name") val targetRootName: String = "",
    val subfolder: String = "",
    val state: String = "",
    val progress: Int = 0,
    @SerialName("speed_bytes_per_second") val speedBytesPerSecond: Long? = null,
    @SerialName("eta_seconds") val etaSeconds: Int? = null,
    @SerialName("file_id") val fileId: Int? = null,
    val error: String = "",
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("ended_at") val endedAt: String? = null,
)
