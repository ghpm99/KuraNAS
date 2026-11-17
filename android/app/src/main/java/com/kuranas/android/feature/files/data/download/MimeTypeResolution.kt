package com.kuranas.android.feature.files.data.download

import android.webkit.MimeTypeMap
import javax.inject.Inject

const val UNKNOWN_MIME_TYPE = "application/octet-stream"

fun interface MimeTypeResolver {
    fun mimeTypeForFileName(fileName: String): String?
}

class AndroidMimeTypeResolver @Inject constructor() : MimeTypeResolver {
    override fun mimeTypeForFileName(fileName: String): String? {
        val extension = fileName.substringAfterLast('.', missingDelimiterValue = "")
        if (extension.isEmpty()) return null
        return MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension.lowercase())
    }
}

/** Extensão do nome primeiro, depois o `Content-Type` da resposta (sem parâmetros), por fim octet-stream. */
fun resolveDownloadMimeType(fileName: String, responseContentType: String?, resolver: MimeTypeResolver): String {
    resolver.mimeTypeForFileName(fileName)?.let { return it }
    val declaredType = responseContentType?.substringBefore(';')?.trim()?.lowercase()
    return declaredType?.takeIf { it.isNotEmpty() } ?: UNKNOWN_MIME_TYPE
}
