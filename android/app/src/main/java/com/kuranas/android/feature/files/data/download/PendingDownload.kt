package com.kuranas.android.feature.files.data.download

import android.net.Uri
import java.io.File
import java.io.FileOutputStream
import java.io.IOException
import java.io.OutputStream

/**
 * Destino de um download que só passa a existir de forma definitiva em [commit].
 * [openOutput] pode ser chamado de novo para retomar: com [shouldAppend] continua o conteúdo já gravado,
 * sem ele recomeça do zero. [discard] remove tudo o que foi gravado.
 * [reference] identifica o destino já criado para que outro processo o reabra; é nulo enquanto nada foi criado.
 */
interface PendingDownload {
    val reference: String?
        get() = null

    @Throws(IOException::class)
    fun openOutput(shouldAppend: Boolean): OutputStream

    @Throws(IOException::class)
    fun commit(): Uri

    fun discard()
}

fun interface PendingDownloadFactory {
    fun create(mimeType: String): PendingDownload
}

class CachedFileDownload(
    private val file: File,
    private val uriForFile: (File) -> Uri,
) : PendingDownload {

    override val reference: String
        get() = file.absolutePath

    override fun openOutput(shouldAppend: Boolean): OutputStream {
        file.parentFile?.mkdirs()
        return FileOutputStream(file, shouldAppend)
    }

    override fun commit(): Uri = uriForFile(file)

    override fun discard() {
        file.delete()
    }
}
