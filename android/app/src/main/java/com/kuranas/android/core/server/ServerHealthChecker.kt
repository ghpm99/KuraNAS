package com.kuranas.android.core.server

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import java.io.IOException
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Confirma que um endereço é um servidor KuraNAS consultando `GET /api/v1/health`.
 * Usa cliente próprio, sem os interceptors que reescrevem o host para o servidor salvo.
 */
@Singleton
class ServerHealthChecker @Inject constructor(json: Json) {

    private val lenientJson = Json(from = json) { ignoreUnknownKeys = true }

    private val httpClient: OkHttpClient = OkHttpClient.Builder()
        .callTimeout(CHECK_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    /**
     * Consulta o health de [baseUrl] (com ou sem barra final).
     * Devolve [ServerCheckResult.Confirmed] só para HTTP 200 com `status=ok` e `service=kuranas`.
     */
    suspend fun check(baseUrl: String): ServerCheckResult = withContext(Dispatchers.IO) {
        val healthUrl = "${baseUrl.trim().trimEnd('/')}$HEALTH_PATH".toHttpUrlOrNull()
            ?: return@withContext ServerCheckResult.NoResponse
        try {
            httpClient.newCall(Request.Builder().url(healthUrl).build()).execute().use(::classify)
        } catch (_: IOException) {
            ServerCheckResult.NoResponse
        }
    }

    private fun classify(response: Response): ServerCheckResult {
        if (response.code == HTTP_FORBIDDEN) return ServerCheckResult.AccessDenied(parseAccessDenial(response.body?.string().orEmpty(), lenientJson))
        if (response.code != HTTP_OK) return ServerCheckResult.NotKuraNas
        val responseText = response.body?.string().orEmpty()
        val payload = runCatching { lenientJson.decodeFromString<HealthPayload>(responseText) }.getOrNull() ?: return ServerCheckResult.NotKuraNas
        val isKuraNas = payload.status == EXPECTED_STATUS && payload.service == EXPECTED_SERVICE
        return if (isKuraNas) ServerCheckResult.Confirmed else ServerCheckResult.NotKuraNas
    }

    @Serializable
    private data class HealthPayload(val status: String? = null, val service: String? = null)

    private companion object {
        const val HEALTH_PATH = "/api/v1/health"
        const val CHECK_TIMEOUT_SECONDS = 3L
        const val HTTP_OK = 200
        const val HTTP_FORBIDDEN = 403
        const val EXPECTED_STATUS = "ok"
        const val EXPECTED_SERVICE = "kuranas"
    }
}
