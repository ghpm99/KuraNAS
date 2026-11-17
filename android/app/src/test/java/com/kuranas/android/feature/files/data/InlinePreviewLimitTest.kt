package com.kuranas.android.feature.files.data

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class InlinePreviewLimitTest {

    private val twentyMegabytes = 20L * 1024 * 1024

    @Test
    fun `files up to 20 MB are previewed inline`() {
        assertTrue(canPreviewInline(0))
        assertTrue(canPreviewInline(twentyMegabytes))
    }

    @Test
    fun `files above 20 MB are not previewed inline`() {
        assertFalse(canPreviewInline(twentyMegabytes + 1))
    }
}
