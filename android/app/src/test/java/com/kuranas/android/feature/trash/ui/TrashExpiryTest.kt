package com.kuranas.android.feature.trash.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import java.time.Instant

class TrashExpiryTest {

    private val currentInstant = Instant.parse("2026-06-11T10:00:00Z")

    @Test
    fun `a partial day counts as a whole one`() {
        assertEquals(2, daysUntilExpiry("2026-06-13T09:00:00Z", currentInstant))
        assertEquals(1, daysUntilExpiry("2026-06-11T11:00:00Z", currentInstant))
    }

    @Test
    fun `an offset timestamp is compared on the same timeline`() {
        assertEquals(1, daysUntilExpiry("2026-06-12T07:00:00-03:00", currentInstant))
    }

    @Test
    fun `an overdue item has zero days left`() {
        assertEquals(0, daysUntilExpiry("2026-06-01T00:00:00Z", currentInstant))
    }

    @Test
    fun `an unreadable or missing expiry is unknown`() {
        assertNull(daysUntilExpiry("", currentInstant))
        assertNull(daysUntilExpiry("amanhã", currentInstant))
    }

    @Test
    fun `origin folder drops the item name`() {
        assertEquals("/data/docs", originFolderOf("/data/docs/a.pdf"))
        assertEquals("C:\\data", originFolderOf("C:\\data\\a.pdf"))
    }

    @Test
    fun `top level items and empty paths come from the root`() {
        assertEquals("/", originFolderOf("/a.pdf"))
        assertEquals("/", originFolderOf(""))
    }
}
