package com.kuranas.android.core.network

import com.kuranas.android.core.server.AccessDeniedEvents
import com.kuranas.android.core.server.parseAccessDenialWithIp
import kotlinx.serialization.json.Json
import okhttp3.Interceptor
import okhttp3.Response

class AccessDenialInterceptor(
    private val accessDeniedEvents: AccessDeniedEvents,
    private val json: Json,
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val response = chain.proceed(chain.request())
        if (response.code == HTTP_FORBIDDEN) reportIfAccessDenied(response)
        return response
    }

    private fun reportIfAccessDenied(response: Response) {
        val denial = runCatching {
            parseAccessDenialWithIp(response.peekBody(PEEK_LIMIT_BYTES).string(), json)
        }.getOrNull() ?: return
        accessDeniedEvents.report(denial)
    }

    private companion object {
        const val HTTP_FORBIDDEN = 403
        const val PEEK_LIMIT_BYTES = 4096L
    }
}
