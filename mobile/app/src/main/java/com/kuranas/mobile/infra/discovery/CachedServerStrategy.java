package com.kuranas.mobile.infra.discovery;

import com.kuranas.mobile.infra.preferences.ServerPreferences;

public final class CachedServerStrategy implements DiscoveryStrategy {

    private final ServerPreferences preferences;
    private final ServerValidator validator;

    public CachedServerStrategy(ServerPreferences preferences, ServerValidator validator) {
        this.preferences = preferences;
        this.validator = validator;
    }

    @Override
    public String name() {
        return "cache";
    }

    @Override
    public void discover(StrategyCallback callback) {
        String cachedUrl = preferences.getServerUrl();
        if (cachedUrl == null || cachedUrl.isEmpty()) {
            callback.onNotFound();
            return;
        }

        if (validator.validate(cachedUrl)) {
            callback.onFound(cachedUrl);
        } else {
            callback.onNotFound();
        }
    }

    @Override
    public void cancel() {
        // No async operation to cancel
    }
}
