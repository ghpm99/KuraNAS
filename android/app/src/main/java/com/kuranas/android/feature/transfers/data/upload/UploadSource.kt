package com.kuranas.android.feature.transfers.data.upload

class UploadSourceUnavailableException(cause: Throwable? = null) : Exception(cause)

/**
 * Origem dos bytes de um envio. Toda função lança [UploadSourceUnavailableException]
 * quando a origem sumiu ou perdeu a permissão de leitura.
 */
interface UploadSource {
    suspend fun sizeBytes(): Long

    suspend fun modifiedAtMillis(): Long?

    suspend fun sha256Hex(): String

    suspend fun readChunk(offset: Long, length: Int): ByteArray
}
