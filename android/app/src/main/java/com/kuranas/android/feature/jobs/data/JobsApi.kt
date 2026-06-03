package com.kuranas.android.feature.jobs.data

import com.kuranas.android.core.network.PageDto
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

interface JobsApi {
    // Sem barra final: a rota de listagem é GET /jobs (envelope de paginação).
    @GET("api/v1/jobs")
    suspend fun listJobs(): PageDto<JobDto>

    @GET("api/v1/jobs/{id}")
    suspend fun getJobById(@Path("id") id: Int): JobDto

    @GET("api/v1/jobs/{id}/steps")
    suspend fun getJobSteps(@Path("id") id: Int): List<JobStepDto>

    @POST("api/v1/jobs/{id}/cancel")
    suspend fun cancelJob(@Path("id") id: Int)
}

/**
 * Espelha jobs.JobDto. `id` é inteiro, `progress` é um objeto e o tipo do job
 * vem em `type`. As propriedades computadas mantêm a API esperada pela UI.
 */
@Serializable
data class JobDto(
    val id: Int = 0,
    val type: String = "",
    val priority: String = "",
    val status: String = "",
    @SerialName("progress") val progressInfo: JobProgressDto = JobProgressDto(),
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("started_at") val startedAt: String? = null,
    @SerialName("ended_at") val endedAt: String? = null,
    @SerialName("cancel_requested") val cancelRequested: Boolean = false,
    @SerialName("last_error") val lastError: String = "",
) {
    val name: String get() = type
    /** Progresso normalizado para 0f..1f a partir do percentual inteiro do backend. */
    val progress: Float get() = progressInfo.progress.coerceIn(0, 100) / 100f
    val error: String? get() = lastError.takeIf { it.isNotBlank() }
    val updatedAt: String get() = endedAt ?: startedAt ?: createdAt
}

@Serializable
data class JobProgressDto(
    @SerialName("total_steps") val totalSteps: Int = 0,
    @SerialName("completed_steps") val completedSteps: Int = 0,
    @SerialName("running_steps") val runningSteps: Int = 0,
    @SerialName("failed_steps") val failedSteps: Int = 0,
    @SerialName("skipped_steps") val skippedSteps: Int = 0,
    @SerialName("canceled_steps") val canceledSteps: Int = 0,
    val progress: Int = 0,
)

/** Espelha jobs.StepDto. */
@Serializable
data class JobStepDto(
    val id: Int = 0,
    @SerialName("job_id") val jobId: Int = 0,
    val type: String = "",
    val status: String = "",
    @SerialName("depends_on") val dependsOn: List<Int> = emptyList(),
    val attempts: Int = 0,
    @SerialName("max_attempts") val maxAttempts: Int = 0,
    @SerialName("last_error") val lastError: String = "",
    val progress: Int = 0,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("started_at") val startedAt: String? = null,
    @SerialName("ended_at") val endedAt: String? = null,
)
