package com.kuranas.android.feature.ingest.data

private val httpUrlPattern = Regex("""https?://[^\s<>"]+""", RegexOption.IGNORE_CASE)
private const val TRAILING_PUNCTUATION = ".,;:!?)]}'\""

/** Primeira URL http/https encontrada no texto compartilhado, sem pontuação final; null se não houver. */
fun firstHttpUrlIn(sharedText: String?): String? {
    val match = httpUrlPattern.find(sharedText.orEmpty()) ?: return null
    val url = match.value.trimEnd { it in TRAILING_PUNCTUATION }
    return url.takeIf { it.length > url.indexOf("://") + 3 }
}
