package com.kuranas.android.core.discovery

import kotlinx.serialization.SerializationException
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
private data class UdpDiscoveryResponse(val service: String? = null, val port: Int? = null)

/** Interpreta a resposta UDP do servidor KuraNAS e monta o servidor descoberto. */
object UdpDiscoveryResponseParser {

    private const val EXPECTED_SERVICE = "kuranas"
    private val lenientJson = Json { ignoreUnknownKeys = true }

    /**
     * Devolve o servidor descrito por [payload], usando [sourceAddress] como host,
     * ou null se o JSON for inválido, o serviço não for "kuranas" ou a porta faltar.
     */
    fun parse(payload: String, sourceAddress: String): DiscoveredServer? {
        val response = decode(payload) ?: return null
        if (response.service != EXPECTED_SERVICE) return null
        val port = response.port ?: return null
        if (port !in 1..65535) return null
        return DiscoveredServer(name = sourceAddress, host = sourceAddress, port = port)
    }

    private fun decode(payload: String): UdpDiscoveryResponse? =
        try {
            lenientJson.decodeFromString<UdpDiscoveryResponse>(payload)
        } catch (_: SerializationException) {
            null
        } catch (_: IllegalArgumentException) {
            null
        }
}
