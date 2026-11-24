package com.kuranas.android.feature.cameraupload.data

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

interface ServerReachability {
    suspend fun isHomeServerReachable(): Boolean
}

fun isHealthyKuranasResponse(statusCode: Int, bodyText: String?, json: Json): Boolean {
    if (statusCode !in HTTP_SUCCESS_RANGE) return false
    val payload = bodyText?.let { runCatching { json.decodeFromString<HealthPayload>(it) }.getOrNull() } ?: return false
    return payload.status == HEALTHY_STATUS
}

private val HTTP_SUCCESS_RANGE = 200..299
private const val HEALTHY_STATUS = "ok"

@Serializable
private data class HealthPayload(val status: String? = null)

@Singleton
class HealthEndpointReachability @Inject constructor(
    okHttpClient: OkHttpClient,
    private val json: Json,
) : ServerReachability {

    private val shortTimeoutClient = okHttpClient.newBuilder()
        .callTimeout(HEALTH_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    override suspend fun isHomeServerReachable(): Boolean = withContext(Dispatchers.IO) {
        try {
            shortTimeoutClient.newCall(Request.Builder().url(HEALTH_URL).build()).execute().use {
                isHealthyKuranasResponse(it.code, it.body?.string(), json)
            }
        } catch (_: IOException) {
            false
        }
    }

    private companion object {
        const val HEALTH_URL = "http://localhost:8000/api/v1/health"
        const val HEALTH_TIMEOUT_SECONDS = 5L
    }
}
