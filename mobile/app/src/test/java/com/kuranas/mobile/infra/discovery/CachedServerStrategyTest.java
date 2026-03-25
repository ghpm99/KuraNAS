package com.kuranas.mobile.infra.discovery;

import com.kuranas.mobile.infra.preferences.ServerPreferences;

import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

public class CachedServerStrategyTest {

    private ServerPreferences mockPreferences;
    private ServerValidator mockValidator;
    private CachedServerStrategy strategy;

    @Before
    public void setUp() {
        mockPreferences = mock(ServerPreferences.class);
        mockValidator = mock(ServerValidator.class);
        strategy = new CachedServerStrategy(mockPreferences, mockValidator);
    }

    @Test
    public void nameReturnsCache() {
        assertEquals("cache", strategy.name());
    }

    @Test
    public void discoverCallsOnNotFoundWhenNoCachedUrl() {
        when(mockPreferences.getServerUrl()).thenReturn(null);

        final boolean[] notFoundCalled = {false};
        strategy.discover(new DiscoveryStrategy.StrategyCallback() {
            @Override
            public void onFound(String serverUrl) {
                throw new AssertionError("Should not be called");
            }

            @Override
            public void onNotFound() {
                notFoundCalled[0] = true;
            }
        });

        assertEquals(true, notFoundCalled[0]);
    }

    @Test
    public void discoverCallsOnNotFoundWhenCachedUrlIsEmpty() {
        when(mockPreferences.getServerUrl()).thenReturn("");

        final boolean[] notFoundCalled = {false};
        strategy.discover(new DiscoveryStrategy.StrategyCallback() {
            @Override
            public void onFound(String serverUrl) {
                throw new AssertionError("Should not be called");
            }

            @Override
            public void onNotFound() {
                notFoundCalled[0] = true;
            }
        });

        assertEquals(true, notFoundCalled[0]);
    }

    @Test
    public void discoverCallsOnFoundWhenCachedUrlIsValid() {
        String cachedUrl = "http://192.168.1.50:8000";
        when(mockPreferences.getServerUrl()).thenReturn(cachedUrl);
        when(mockValidator.validate(cachedUrl)).thenReturn(true);

        final String[] foundUrl = {null};
        strategy.discover(new DiscoveryStrategy.StrategyCallback() {
            @Override
            public void onFound(String serverUrl) {
                foundUrl[0] = serverUrl;
            }

            @Override
            public void onNotFound() {
                throw new AssertionError("Should not be called");
            }
        });

        assertEquals(cachedUrl, foundUrl[0]);
    }

    @Test
    public void discoverCallsOnNotFoundWhenCachedUrlFailsValidation() {
        String cachedUrl = "http://192.168.1.50:8000";
        when(mockPreferences.getServerUrl()).thenReturn(cachedUrl);
        when(mockValidator.validate(cachedUrl)).thenReturn(false);

        final boolean[] notFoundCalled = {false};
        strategy.discover(new DiscoveryStrategy.StrategyCallback() {
            @Override
            public void onFound(String serverUrl) {
                throw new AssertionError("Should not be called");
            }

            @Override
            public void onNotFound() {
                notFoundCalled[0] = true;
            }
        });

        assertEquals(true, notFoundCalled[0]);
    }
}
