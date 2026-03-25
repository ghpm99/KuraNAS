package com.kuranas.mobile.infra.discovery;

import android.util.Log;

import com.kuranas.mobile.domain.model.DiscoveryResult;
import com.kuranas.mobile.domain.port.ServerDiscoveryPort;
import com.kuranas.mobile.infra.preferences.ServerPreferences;

import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

public final class ServerDiscovery implements ServerDiscoveryPort {

    private static final String TAG = "ServerDiscovery";

    private final List<DiscoveryStrategy> strategies;
    private final ServerPreferences preferences;
    private final ExecutorService executor;
    private volatile boolean cancelled;
    private volatile DiscoveryStrategy currentStrategy;

    public ServerDiscovery(List<DiscoveryStrategy> strategies, ServerPreferences preferences) {
        this.strategies = strategies;
        this.preferences = preferences;
        this.executor = Executors.newSingleThreadExecutor();
    }

    @Override
    public void discover(final DiscoveryCallback callback) {
        cancelled = false;

        executor.submit(new Runnable() {
            @Override
            public void run() {
                for (final DiscoveryStrategy strategy : strategies) {
                    if (cancelled) break;

                    currentStrategy = strategy;
                    Log.d(TAG, "Trying strategy: " + strategy.name());
                    callback.onProgress(strategy.name());

                    final CountDownLatch latch = new CountDownLatch(1);
                    final AtomicReference<String> foundUrl = new AtomicReference<String>(null);

                    strategy.discover(new DiscoveryStrategy.StrategyCallback() {
                        @Override
                        public void onFound(String serverUrl) {
                            foundUrl.set(serverUrl);
                            latch.countDown();
                        }

                        @Override
                        public void onNotFound() {
                            latch.countDown();
                        }
                    });

                    try {
                        latch.await(30, TimeUnit.SECONDS);
                    } catch (InterruptedException e) {
                        Log.d(TAG, "Strategy interrupted: " + strategy.name());
                        break;
                    }

                    String url = foundUrl.get();
                    if (url != null && !cancelled) {
                        Log.d(TAG, "Server found via " + strategy.name() + ": " + url);
                        preferences.saveServerUrl(url);
                        callback.onDiscovered(new DiscoveryResult(url, strategy.name()));
                        return;
                    }

                    Log.d(TAG, "Strategy " + strategy.name() + " did not find server");
                }

                if (!cancelled) {
                    callback.onFailed("No server found");
                }
            }
        });
    }

    @Override
    public void cancel() {
        cancelled = true;
        DiscoveryStrategy strategy = currentStrategy;
        if (strategy != null) {
            strategy.cancel();
        }
    }
}
