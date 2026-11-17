package com.kuranas.android.feature.files.data

import android.content.ContentResolver
import android.net.Uri
import okhttp3.MediaType
import okhttp3.RequestBody
import okio.BufferedSink
import okio.source
import java.io.FileNotFoundException

/**
 * Corpo de requisição que lê o conteúdo de [uri] por fluxo apenas quando a requisição é escrita,
 * sem carregar o arquivo na memória. [sizeBytes] negativo indica tamanho desconhecido.
 */
class UriRequestBody(
    private val contentResolver: ContentResolver,
    private val uri: Uri,
    private val mediaType: MediaType?,
    private val sizeBytes: Long,
) : RequestBody() {

    override fun contentType(): MediaType? = mediaType

    override fun contentLength(): Long = sizeBytes

    override fun writeTo(sink: BufferedSink) {
        val stream = contentResolver.openInputStream(uri) ?: throw FileNotFoundException(uri.toString())
        stream.source().use { source -> sink.writeAll(source) }
    }
}
