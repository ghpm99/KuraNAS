package com.kuranas.android.feature.transfers.data.upload

import android.content.ContentResolver
import android.net.Uri
import android.os.ParcelFileDescriptor
import android.provider.DocumentsContract
import android.provider.OpenableColumns
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.FileInputStream
import java.io.FileNotFoundException
import java.nio.ByteBuffer

class ContentUriUploadSource(
    private val contentResolver: ContentResolver,
    private val uri: Uri,
) : UploadSource {

    override suspend fun sizeBytes(): Long = withContext(Dispatchers.IO) {
        guarded {
            querySizeColumn() ?: contentResolver.openFileDescriptor(uri, READ_MODE).use { descriptor ->
                descriptor?.statSize ?: throw FileNotFoundException(uri.toString())
            }
        }
    }

    override suspend fun modifiedAtMillis(): Long? = withContext(Dispatchers.IO) {
        runCatching {
            contentResolver.query(uri, arrayOf(DocumentsContract.Document.COLUMN_LAST_MODIFIED), null, null, null)
                ?.use { cursor ->
                    if (!cursor.moveToFirst() || cursor.isNull(0)) return@use null
                    cursor.getLong(0).takeIf { it > 0 }
                }
        }.getOrNull()
    }

    override suspend fun sha256Hex(): String = withContext(Dispatchers.IO) {
        guarded {
            val stream = contentResolver.openInputStream(uri) ?: throw FileNotFoundException(uri.toString())
            stream.use { Sha256Hasher.hexOf(it) }
        }
    }

    override suspend fun readChunk(offset: Long, length: Int): ByteArray = withContext(Dispatchers.IO) {
        guarded {
            val descriptor = contentResolver.openFileDescriptor(uri, READ_MODE)
                ?: throw FileNotFoundException(uri.toString())
            descriptor.use { openDescriptor -> readRange(openDescriptor, offset, length) }
        }
    }

    private fun readRange(descriptor: ParcelFileDescriptor, offset: Long, length: Int): ByteArray {
        FileInputStream(descriptor.fileDescriptor).channel.use { channel ->
            channel.position(offset)
            val buffer = ByteBuffer.allocate(length)
            while (buffer.hasRemaining()) {
                if (channel.read(buffer) < 0) break
            }
            return buffer.array().copyOf(buffer.position())
        }
    }

    private fun querySizeColumn(): Long? =
        contentResolver.query(uri, arrayOf(OpenableColumns.SIZE), null, null, null)?.use { cursor ->
            if (!cursor.moveToFirst() || cursor.isNull(0)) return@use null
            cursor.getLong(0).takeIf { it >= 0 }
        }

    private fun <T> guarded(block: () -> T): T =
        try {
            block()
        } catch (exception: FileNotFoundException) {
            throw UploadSourceUnavailableException(exception)
        } catch (exception: SecurityException) {
            throw UploadSourceUnavailableException(exception)
        }

    private companion object {
        const val READ_MODE = "r"
    }
}
