package com.kuranas.mobile.domain.model;

import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;
import static org.junit.Assert.assertNotNull;

public class DiscoveryResultTest {

    @Test
    public void constructorSetsFields() {
        DiscoveryResult result = new DiscoveryResult("http://192.168.1.50:8000", "cache");

        assertEquals("http://192.168.1.50:8000", result.getServerUrl());
        assertEquals("cache", result.getSource());
    }

    @Test(expected = IllegalArgumentException.class)
    public void constructorRejectsNullUrl() {
        new DiscoveryResult(null, "cache");
    }

    @Test(expected = IllegalArgumentException.class)
    public void constructorRejectsNullSource() {
        new DiscoveryResult("http://192.168.1.50:8000", null);
    }

    @Test
    public void equalObjectsAreEqual() {
        DiscoveryResult a = new DiscoveryResult("http://192.168.1.50:8000", "mdns");
        DiscoveryResult b = new DiscoveryResult("http://192.168.1.50:8000", "mdns");

        assertEquals(a, b);
        assertEquals(a.hashCode(), b.hashCode());
    }

    @Test
    public void differentUrlsAreNotEqual() {
        DiscoveryResult a = new DiscoveryResult("http://192.168.1.50:8000", "mdns");
        DiscoveryResult b = new DiscoveryResult("http://192.168.1.51:8000", "mdns");

        assertNotEquals(a, b);
    }

    @Test
    public void differentSourcesAreNotEqual() {
        DiscoveryResult a = new DiscoveryResult("http://192.168.1.50:8000", "mdns");
        DiscoveryResult b = new DiscoveryResult("http://192.168.1.50:8000", "udp");

        assertNotEquals(a, b);
    }

    @Test
    public void toStringContainsFields() {
        DiscoveryResult result = new DiscoveryResult("http://192.168.1.50:8000", "scan");
        String str = result.toString();

        assertNotNull(str);
        assertEquals(true, str.contains("192.168.1.50"));
        assertEquals(true, str.contains("scan"));
    }
}
