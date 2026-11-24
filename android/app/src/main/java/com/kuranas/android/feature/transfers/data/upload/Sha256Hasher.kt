package com.kuranas.android.feature.transfers.data.upload

import java.io.InputStream
import java.security.MessageDigest

internal object Sha256Hasher {
    private const val BUFFER_SIZE_BYTES = 64 * 1024

    fun hexOf(stream: InputStream): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val buffer = ByteArray(BUFFER_SIZE_BYTES)
        while (true) {
            val read = stream.read(buffer)
            if (read < 0) break
            digest.update(buffer, 0, read)
        }
        return digest.digest().joinToString("") { "%02x".format(it) }
    }
}
