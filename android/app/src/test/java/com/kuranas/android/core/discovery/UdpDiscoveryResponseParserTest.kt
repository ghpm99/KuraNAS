package com.kuranas.android.core.discovery

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class UdpDiscoveryResponseParserTest {

    @Test
    fun `valid response uses source address and response port`() {
        val server = UdpDiscoveryResponseParser.parse(
            """{"service":"kuranas","port":8123,"api":"/api/v1"}""",
            "192.168.0.10",
        )
        assertEquals("http://192.168.0.10:8123", server?.url)
    }

    @Test
    fun `other service is ignored`() {
        assertNull(UdpDiscoveryResponseParser.parse("""{"service":"other","port":8000}""", "192.168.0.10"))
    }

    @Test
    fun `invalid json is ignored`() {
        assertNull(UdpDiscoveryResponseParser.parse("not json", "192.168.0.10"))
    }

    @Test
    fun `missing port is ignored`() {
        assertNull(UdpDiscoveryResponseParser.parse("""{"service":"kuranas"}""", "192.168.0.10"))
    }
}
