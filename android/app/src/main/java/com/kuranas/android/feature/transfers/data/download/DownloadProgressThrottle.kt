package com.kuranas.android.feature.transfers.data.download

import com.kuranas.android.feature.files.data.download.DownloadProgress

class DownloadProgressThrottle(
    private val minimumBytesBetweenPersists: Long = DEFAULT_MINIMUM_BYTES,
) {

    private var lastPersistedPercent: Int? = null
    private var lastPersistedBytes: Long? = null

    fun shouldPersist(progress: DownloadProgress): Boolean {
        val percent = progress.fraction?.let { (it * PERCENT_SCALE).toInt() }
        val bytesSincePersist = progress.receivedBytes - (lastPersistedBytes ?: 0)
        val isFirst = lastPersistedBytes == null
        val hasPercentChanged = percent != null && percent != lastPersistedPercent
        val hasMovedBackwards = bytesSincePersist < 0
        val isDue = isFirst || hasPercentChanged || hasMovedBackwards || bytesSincePersist >= minimumBytesBetweenPersists
        if (!isDue) return false
        lastPersistedPercent = percent
        lastPersistedBytes = progress.receivedBytes
        return true
    }

    private companion object {
        const val DEFAULT_MINIMUM_BYTES = 4L * 1024 * 1024
        const val PERCENT_SCALE = 100
    }
}
