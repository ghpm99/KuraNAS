package com.kuranas.mobile.infra.discovery;

import com.kuranas.mobile.domain.model.DiscoveryResult;
import com.kuranas.mobile.domain.port.ServerDiscoveryPort;
import com.kuranas.mobile.infra.preferences.ServerPreferences;

import org.junit.Before;
import org.junit.Test;

import java.util.Arrays;
import java.util.Collections;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

public class ServerDiscoveryTest {

    private ServerPreferences mockPreferences;

    @Before
    public void setUp() {
        mockPreferences = mock(ServerPreferences.class);
    }

    private DiscoveryStrategy successStrategy(final String name, final String url) {
        return new DiscoveryStrategy() {
            @Override
            public String name() { return name; }

            @Override
            public void discover(StrategyCallback callback) {
                callback.onFound(url);
            }

            @Override
            public void cancel() {}
        };
    }

    private DiscoveryStrategy failStrategy(final String name) {
        return new DiscoveryStrategy() {
            @Override
            public String name() { return name; }

            @Override
            public void discover(StrategyCallback callback) {
                callback.onNotFound();
            }

            @Override
            public void cancel() {}
        };
    }

    @Test
    public void discoveryStopsAtFirstSuccess() throws Exception {
        String expectedUrl = "http://192.168.1.50:8000";
        ServerDiscovery discovery = new ServerDiscovery(
                Arrays.asList(
                        failStrategy("cache"),
                        successStrategy("udp", expectedUrl),
                        failStrategy("scan")
                ),
                mockPreferences
        );

        final CountDownLatch latch = new CountDownLatch(1);
        final DiscoveryResult[] result = {null};

        discovery.discover(new ServerDiscoveryPort.DiscoveryCallback() {
            @Override
            public void onDiscovered(DiscoveryResult r) {
                result[0] = r;
                latch.countDown();
            }

            @Override
            public void onFailed(String reason) {
                latch.countDown();
            }

            @Override
            public void onProgress(String strategyName) {}
        });

        latch.await(5, TimeUnit.SECONDS);

        assertNotNull(result[0]);
        assertEquals(expectedUrl, result[0].getServerUrl());
        assertEquals("udp", result[0].getSource());
        verify(mockPreferences).saveServerUrl(expectedUrl);
    }

    @Test
    public void discoveryReportsFailureWhenAllStrategiesFail() throws Exception {
        ServerDiscovery discovery = new ServerDiscovery(
                Arrays.asList(
                        failStrategy("cache"),
                        failStrategy("udp"),
                        failStrategy("scan")
                ),
                mockPreferences
        );

        final CountDownLatch latch = new CountDownLatch(1);
        final String[] failReason = {null};

        discovery.discover(new ServerDiscoveryPort.DiscoveryCallback() {
            @Override
            public void onDiscovered(DiscoveryResult r) {
                latch.countDown();
            }

            @Override
            public void onFailed(String reason) {
                failReason[0] = reason;
                latch.countDown();
            }

            @Override
            public void onProgress(String strategyName) {}
        });

        latch.await(5, TimeUnit.SECONDS);

        assertNotNull(failReason[0]);
    }

    @Test
    public void discoveryReportsProgressForEachStrategy() throws Exception {
        ServerDiscovery discovery = new ServerDiscovery(
                Arrays.asList(
                        failStrategy("cache"),
                        failStrategy("mdns"),
                        successStrategy("udp", "http://192.168.1.50:8000")
                ),
                mockPreferences
        );

        final CountDownLatch latch = new CountDownLatch(1);
        final java.util.List<String> progressNames = Collections.synchronizedList(
                new java.util.ArrayList<String>()
        );

        discovery.discover(new ServerDiscoveryPort.DiscoveryCallback() {
            @Override
            public void onDiscovered(DiscoveryResult r) {
                latch.countDown();
            }

            @Override
            public void onFailed(String reason) {
                latch.countDown();
            }

            @Override
            public void onProgress(String strategyName) {
                progressNames.add(strategyName);
            }
        });

        latch.await(5, TimeUnit.SECONDS);

        assertEquals(3, progressNames.size());
        assertEquals("cache", progressNames.get(0));
        assertEquals("mdns", progressNames.get(1));
        assertEquals("udp", progressNames.get(2));
    }

    @Test
    public void discoveryWithEmptyStrategiesReportsFailed() throws Exception {
        ServerDiscovery discovery = new ServerDiscovery(
                Collections.<DiscoveryStrategy>emptyList(),
                mockPreferences
        );

        final CountDownLatch latch = new CountDownLatch(1);
        final String[] failReason = {null};

        discovery.discover(new ServerDiscoveryPort.DiscoveryCallback() {
            @Override
            public void onDiscovered(DiscoveryResult r) {
                latch.countDown();
            }

            @Override
            public void onFailed(String reason) {
                failReason[0] = reason;
                latch.countDown();
            }

            @Override
            public void onProgress(String strategyName) {}
        });

        latch.await(5, TimeUnit.SECONDS);

        assertNotNull(failReason[0]);
    }
}
