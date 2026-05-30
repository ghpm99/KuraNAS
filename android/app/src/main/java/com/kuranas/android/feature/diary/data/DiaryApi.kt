package com.kuranas.android.feature.diary.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

interface DiaryApi {
    @GET("api/v1/diary/")
    suspend fun getDiary(): List<DiaryEntryDto>

    @GET("api/v1/diary/summary")
    suspend fun getSummary(): DiarySummaryDto

    @POST("api/v1/diary/")
    suspend fun createEntry(@Body body: CreateDiaryRequest): DiaryEntryDto

    @PUT("api/v1/diary/{id}")
    suspend fun updateEntry(@Path("id") id: Int, @Body body: CreateDiaryRequest): DiaryEntryDto

    @POST("api/v1/diary/copy")
    suspend fun duplicateEntry(@Body body: DuplicateDiaryRequest): DiaryEntryDto
}

@Serializable
data class DiaryEntryDto(
    val id: Int = 0,
    val title: String = "",
    val content: String = "",
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
)

@Serializable
data class DiarySummaryDto(val total: Int = 0, @SerialName("recent_count") val recentCount: Int = 0)

@Serializable
data class CreateDiaryRequest(val title: String, val content: String)

@Serializable
data class DuplicateDiaryRequest(val id: Int)
