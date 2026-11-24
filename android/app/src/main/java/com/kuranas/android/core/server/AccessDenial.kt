package com.kuranas.android.core.server

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

/** Recusa do servidor a este aparelho: IP visto pelo servidor e a mensagem já traduzida por ele. */
data class AccessDenial(val deviceIp: String?, val serverMessage: String?)

@Serializable
private data class AccessDeniedBody(val error: String? = null, val ip: String? = null)

/** Lê o corpo de um 403 da whitelist; devolve null quando o corpo não traz o campo `ip`. */
fun parseAccessDenialWithIp(body: String, json: Json): AccessDenial? {
    val parsedBody = runCatching { json.decodeFromString<AccessDeniedBody>(body) }.getOrNull() ?: return null
    val deviceIp = parsedBody.ip?.takeIf { it.isNotBlank() } ?: return null
    return AccessDenial(deviceIp, parsedBody.error?.takeIf { it.isNotBlank() })
}

/** Lê o corpo de um 403 do health; sem `ip` legível ainda é uma recusa, com IP desconhecido. */
fun parseAccessDenial(body: String, json: Json): AccessDenial =
    parseAccessDenialWithIp(body, json) ?: AccessDenial(deviceIp = null, serverMessage = null)
