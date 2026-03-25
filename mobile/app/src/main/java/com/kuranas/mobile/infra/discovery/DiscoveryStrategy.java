package com.kuranas.mobile.infra.discovery;

public interface DiscoveryStrategy {

    String name();

    void discover(StrategyCallback callback);

    void cancel();

    interface StrategyCallback {
        void onFound(String serverUrl);
        void onNotFound();
    }
}
