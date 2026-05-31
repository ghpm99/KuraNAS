package com.kuranas.android.core.network

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Envelope de paginação padrão do backend (utils.PaginationResponse[T]):
 * { "items": [...], "pagination": { ... } }
 */
@Serializable
data class PageDto<T>(
    val items: List<T> = emptyList(),
    val pagination: PaginationDto = PaginationDto(),
)

@Serializable
data class PaginationDto(
    val page: Int = 1,
    @SerialName("page_size") val pageSize: Int = 0,
    @SerialName("has_next") val hasNext: Boolean = false,
    @SerialName("has_prev") val hasPrev: Boolean = false,
)
