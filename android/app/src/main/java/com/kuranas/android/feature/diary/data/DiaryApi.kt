package com.kuranas.android.feature.diary.data

import com.kuranas.android.core.network.PageDto
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.Field
import retrofit2.http.FormUrlEncoded
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

interface DiaryApi {
    // O backend devolve o envelope de paginação (utils.PaginationResponse), não um array.
    @GET("api/v1/diary/")
    suspend fun getDiary(): PageDto<DiaryEntryDto>

    @GET("api/v1/diary/summary")
    suspend fun getSummary(): DiarySummaryDto

    @POST("api/v1/diary/")
    suspend fun createEntry(@Body body: CreateDiaryRequest): DiaryEntryDto

    // O UpdateDiaryHandler lê o campo de formulário `data` (PostForm), não JSON,
    // e só atualiza o nome da atividade.
    @FormUrlEncoded
    @PUT("api/v1/diary/{id}")
    suspend fun updateEntry(@Path("id") id: Int, @Field("data") name: String): DiaryEntryDto

    @POST("api/v1/diary/copy")
    suspend fun duplicateEntry(@Body body: DuplicateDiaryRequest): DiaryEntryDto
}

/**
 * Espelha diary.DiaryDto (registro de atividade: name/description/start_time/...).
 * As propriedades computadas mantêm a API de "nota" esperada pela tela.
 */
@Serializable
data class DiaryEntryDto(
    val id: Int = 0,
    val name: String = "",
    val description: String = "",
    @SerialName("start_time") val startTime: String = "",
    val duration: Int = 0,
) {
    val title: String get() = name
    val content: String get() = description
    val createdAt: String get() = startTime
}

/** Espelha diary.DiarySummary. */
@Serializable
data class DiarySummaryDto(
    val date: String = "",
    @SerialName("total_activities") val totalActivities: Int = 0,
    @SerialName("total_time_spent_seconds") val totalTimeSpentSeconds: Int = 0,
    @SerialName("longest_activity") val longestActivity: LongestActivityDto? = null,
)

@Serializable
data class LongestActivityDto(
    val name: String = "",
    @SerialName("duration_seconds") val durationSeconds: Int = 0,
    @SerialName("duration_formatted") val durationFormatted: String = "",
)

@Serializable
data class CreateDiaryRequest(val name: String, val description: String = "")

@Serializable
data class DuplicateDiaryRequest(val id: Int)
