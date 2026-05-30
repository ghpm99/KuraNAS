package com.kuranas.android.feature.jobs.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

interface JobsApi {
    @GET("api/v1/jobs/")
    suspend fun listJobs(): List<JobDto>

    @GET("api/v1/jobs/{id}")
    suspend fun getJobById(@Path("id") id: String): JobDto

    @GET("api/v1/jobs/{id}/steps")
    suspend fun getJobSteps(@Path("id") id: String): List<JobStepDto>

    @POST("api/v1/jobs/{id}/cancel")
    suspend fun cancelJob(@Path("id") id: String)
}

@Serializable
data class JobDto(
    val id: String = "",
    val name: String = "",
    val status: String = "",
    val progress: Float = 0f,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    val error: String? = null,
)

@Serializable
data class JobStepDto(
    val id: String = "",
    val name: String = "",
    val status: String = "",
    val message: String = "",
)
