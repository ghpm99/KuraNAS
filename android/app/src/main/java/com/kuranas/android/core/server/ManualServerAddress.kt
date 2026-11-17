package com.kuranas.android.core.server

import okhttp3.HttpUrl.Companion.toHttpUrlOrNull

/** Normaliza o endereço digitado pelo dono para a URL base do servidor. */
object ManualServerAddress {

    private const val DEFAULT_PORT = 8000
    private const val DEFAULT_SCHEME = "http"
    private const val SCHEME_SEPARATOR = "://"
    private val acceptedSchemes = setOf("http", "https")

    /**
     * Devolve `esquema://host:porta` sem path, ou null quando o texto não é um endereço válido.
     * Sem esquema assume http; sem porta assume 8000.
     */
    fun normalize(typedAddress: String): String? {
        val trimmedAddress = typedAddress.trim()
        if (trimmedAddress.isEmpty()) return null

        val scheme = trimmedAddress.substringBefore(SCHEME_SEPARATOR, DEFAULT_SCHEME).lowercase()
        if (scheme !in acceptedSchemes) return null

        val authority = trimmedAddress
            .substringAfter(SCHEME_SEPARATOR, trimmedAddress)
            .substringBefore('/')
            .substringBefore('?')
            .substringBefore('#')
        if (authority.isEmpty()) return null

        val parsedUrl = "$scheme$SCHEME_SEPARATOR$authority".toHttpUrlOrNull() ?: return null
        val port = if (hasExplicitPort(authority)) parsedUrl.port else DEFAULT_PORT
        val host = if (parsedUrl.host.contains(':')) "[${parsedUrl.host}]" else parsedUrl.host
        return "$scheme$SCHEME_SEPARATOR$host:$port"
    }

    private fun hasExplicitPort(authority: String): Boolean =
        authority.substringAfterLast(']').contains(':')
}
