package com.kuranas.mobile.domain.port;

import com.kuranas.mobile.domain.model.DiscoveryResult;

public interface ServerDiscoveryPort {

    void discover(DiscoveryCallback callback);

    void cancel();

    interface DiscoveryCallback {
        void onDiscovered(DiscoveryResult result);
        void onFailed(String reason);
        void onProgress(String strategyName);
    }
}
