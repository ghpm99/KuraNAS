package com.kuranas.android.feature.transfers.data.download

import com.kuranas.android.feature.files.data.download.DownloadProgress
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class DownloadProgressThrottleTest {

    private val throttle = DownloadProgressThrottle(minimumBytesBetweenPersists = 1_000)

    @Test
    fun persistsTheFirstReportAndOnlyWhenThePercentChanges() {
        assertTrue(throttle.shouldPersist(DownloadProgress(0, 100_000)))
        assertFalse(throttle.shouldPersist(DownloadProgress(500, 100_000)))
        assertTrue(throttle.shouldPersist(DownloadProgress(1_000, 100_000)))
    }

    @Test
    fun withoutAKnownTotalPersistsEveryMinimumBytes() {
        assertTrue(throttle.shouldPersist(DownloadProgress(10, -1)))
        assertFalse(throttle.shouldPersist(DownloadProgress(500, -1)))
        assertTrue(throttle.shouldPersist(DownloadProgress(1_010, -1)))
    }

    @Test
    fun persistsWhenTheDownloadRestartsFromZero() {
        assertTrue(throttle.shouldPersist(DownloadProgress(50_000, 1_000_000)))
        assertTrue(throttle.shouldPersist(DownloadProgress(100, 1_000_000)))
    }
}
