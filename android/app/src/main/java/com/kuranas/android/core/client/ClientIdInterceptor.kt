package com.kuranas.android.core.client

import okhttp3.Interceptor
import okhttp3.Response

const val CLIENT_ID_HEADER = "X-KuraNAS-Client-Id"

private val playerPathSegments = listOf("/api/v1/video/", "/api/v1/music/")

class ClientIdInterceptor(private val clientIdProvider: ClientIdProvider) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        if (!isPlayerPath(request.url.encodedPath)) return chain.proceed(request)
        val requestWithClientId = request.newBuilder()
            .header(CLIENT_ID_HEADER, clientIdProvider.clientId())
            .build()
        return chain.proceed(requestWithClientId)
    }

    private fun isPlayerPath(encodedPath: String): Boolean =
        playerPathSegments.any { encodedPath.contains(it) }
}
