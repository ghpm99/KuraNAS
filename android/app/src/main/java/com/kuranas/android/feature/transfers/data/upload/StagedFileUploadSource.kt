package com.kuranas.android.feature.transfers.data.upload

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.io.FileNotFoundException
import java.io.RandomAccessFile

class StagedFileUploadSource(private val file: File) : UploadSource {

    override suspend fun sizeBytes(): Long = withContext(Dispatchers.IO) {
        requireReadable()
        file.length()
    }

    override suspend fun modifiedAtMillis(): Long? = withContext(Dispatchers.IO) {
        file.lastModified().takeIf { it > 0 }
    }

    override suspend fun sha256Hex(): String = withContext(Dispatchers.IO) {
        requireReadable()
        file.inputStream().use { Sha256Hasher.hexOf(it) }
    }

    override suspend fun readChunk(offset: Long, length: Int): ByteArray = withContext(Dispatchers.IO) {
        requireReadable()
        try {
            RandomAccessFile(file, "r").use { access ->
                access.seek(offset)
                val buffer = ByteArray(length)
                var filled = 0
                while (filled < length) {
                    val read = access.read(buffer, filled, length - filled)
                    if (read < 0) break
                    filled += read
                }
                buffer.copyOf(filled)
            }
        } catch (exception: FileNotFoundException) {
            throw UploadSourceUnavailableException(exception)
        }
    }

    private fun requireReadable() {
        if (!file.isFile || !file.canRead()) throw UploadSourceUnavailableException()
    }
}
