package com.kuranas.android.feature.notifications.data

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.GET
import retrofit2.http.PUT
import retrofit2.http.Path

interface NotificationsApi {
    @GET("api/v1/notifications/")
    suspend fun listNotifications(): List<NotificationDto>

    @GET("api/v1/notifications/unread-count")
    suspend fun getUnreadCount(): UnreadCountDto

    @PUT("api/v1/notifications/{id}/read")
    suspend fun markAsRead(@Path("id") id: Int)

    @PUT("api/v1/notifications/read-all")
    suspend fun markAllAsRead()
}

@Serializable
data class NotificationDto(
    val id: Int = 0,
    val title: String = "",
    val message: String = "",
    @SerialName("is_read") val isRead: Boolean = false,
    @SerialName("created_at") val createdAt: String = "",
    val type: String = "",
)

@Serializable
data class UnreadCountDto(val count: Int = 0)
