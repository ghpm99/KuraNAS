package com.kuranas.android.feature.ingest.data

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class IngestDownloadStateTest {

    private fun downloadIn(state: String, jobId: Int? = 7) = IngestDownloadDto(id = 1, jobId = jobId, state = state)

    @Test
    fun `queued and running downloads can be canceled`() {
        assertTrue(downloadIn("queued").canBeCanceled())
        assertTrue(downloadIn("running").canBeCanceled())
    }

    @Test
    fun `finished downloads cannot be canceled`() {
        assertFalse(downloadIn("completed").canBeCanceled())
        assertFalse(downloadIn("failed").canBeCanceled())
    }

    @Test
    fun `an active download without a job cannot be canceled`() {
        assertFalse(downloadIn("running", jobId = null).canBeCanceled())
    }

    @Test
    fun `only failed and canceled downloads can be retried`() {
        assertTrue(downloadIn("failed").canBeRetried())
        assertTrue(downloadIn("canceled").canBeRetried())
        assertFalse(downloadIn("running").canBeRetried())
        assertFalse(downloadIn("completed").canBeRetried())
    }
}
