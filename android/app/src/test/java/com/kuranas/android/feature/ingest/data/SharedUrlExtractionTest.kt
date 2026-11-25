package com.kuranas.android.feature.ingest.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SharedUrlExtractionTest {

    @Test
    fun `returns the url when the whole text is a url`() {
        assertEquals("https://youtu.be/abc?t=5", firstHttpUrlIn("https://youtu.be/abc?t=5"))
    }

    @Test
    fun `returns the first url inside surrounding text`() {
        val sharedText = "Veja isso http://example.com/a e também https://other.com/b"

        assertEquals("http://example.com/a", firstHttpUrlIn(sharedText))
    }

    @Test
    fun `drops trailing punctuation from the url`() {
        assertEquals("https://example.com/video", firstHttpUrlIn("Olha (https://example.com/video)."))
    }

    @Test
    fun `accepts an uppercase scheme`() {
        assertEquals("HTTPS://example.com/x", firstHttpUrlIn("HTTPS://example.com/x"))
    }

    @Test
    fun `returns null when the text has no url`() {
        assertNull(firstHttpUrlIn("apenas um texto qualquer"))
    }

    @Test
    fun `returns null for a non http scheme`() {
        assertNull(firstHttpUrlIn("ftp://example.com/file e file:///sdcard/a.txt"))
    }

    @Test
    fun `returns null for a scheme without host`() {
        assertNull(firstHttpUrlIn("https://"))
    }

    @Test
    fun `returns null for null or empty text`() {
        assertNull(firstHttpUrlIn(null))
        assertNull(firstHttpUrlIn(""))
    }
}
