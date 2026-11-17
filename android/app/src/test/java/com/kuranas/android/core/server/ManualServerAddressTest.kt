package com.kuranas.android.core.server

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ManualServerAddressTest {

    @Test
    fun `adds http scheme and default port`() {
        assertEquals("http://192.168.0.10:8000", ManualServerAddress.normalize("192.168.0.10"))
    }

    @Test
    fun `keeps explicit port`() {
        assertEquals("http://nas.local:9000", ManualServerAddress.normalize("nas.local:9000"))
    }

    @Test
    fun `https without port gets default port`() {
        assertEquals("https://x:8000", ManualServerAddress.normalize("https://x"))
    }

    @Test
    fun `drops path and trailing slash`() {
        assertEquals("http://x:8000", ManualServerAddress.normalize("http://x:8000/api/"))
    }

    @Test
    fun `trims surrounding spaces`() {
        assertEquals("http://x:8000", ManualServerAddress.normalize("  x  "))
    }

    @Test
    fun `rejects unsupported scheme`() {
        assertNull(ManualServerAddress.normalize("ftp://x"))
    }

    @Test
    fun `rejects blank input`() {
        assertNull(ManualServerAddress.normalize(""))
        assertNull(ManualServerAddress.normalize("   "))
    }

    @Test
    fun `rejects invalid host`() {
        assertNull(ManualServerAddress.normalize("http://"))
        assertNull(ManualServerAddress.normalize("bad host"))
    }
}
