package com.kuranas.mobile.infra.discovery;

import android.content.Context;
import android.net.nsd.NsdManager;
import android.net.nsd.NsdServiceInfo;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

public final class NsdDiscoveryStrategy implements DiscoveryStrategy {

    private static final String TAG = "NsdDiscovery";
    private static final String SERVICE_TYPE = "_kuranas._tcp.";
    private static final long DISCOVERY_TIMEOUT_MS = 5000;

    private final Context context;
    private final ServerValidator validator;
    private NsdManager nsdManager;
    private NsdManager.DiscoveryListener discoveryListener;
    private volatile boolean cancelled;
    private volatile boolean completed;

    public NsdDiscoveryStrategy(Context context, ServerValidator validator) {
        this.context = context.getApplicationContext();
        this.validator = validator;
    }

    @Override
    public String name() {
        return "mdns";
    }

    @Override
    public void discover(final StrategyCallback callback) {
        cancelled = false;
        completed = false;

        try {
            nsdManager = (NsdManager) context.getSystemService(Context.NSD_SERVICE);
        } catch (Exception e) {
            Log.e(TAG, "NsdManager not available", e);
            callback.onNotFound();
            return;
        }

        if (nsdManager == null) {
            callback.onNotFound();
            return;
        }

        final Handler timeoutHandler = new Handler(Looper.getMainLooper());
        final Runnable timeoutRunnable = new Runnable() {
            @Override
            public void run() {
                if (!completed && !cancelled) {
                    completed = true;
                    stopDiscovery();
                    callback.onNotFound();
                }
            }
        };

        discoveryListener = new NsdManager.DiscoveryListener() {
            @Override
            public void onDiscoveryStarted(String serviceType) {
                Log.d(TAG, "Discovery started for " + serviceType);
            }

            @Override
            public void onServiceFound(NsdServiceInfo serviceInfo) {
                if (cancelled || completed) return;

                Log.d(TAG, "Service found: " + serviceInfo.getServiceName());
                resolveService(serviceInfo, callback, timeoutHandler, timeoutRunnable);
            }

            @Override
            public void onServiceLost(NsdServiceInfo serviceInfo) {
                Log.d(TAG, "Service lost: " + serviceInfo.getServiceName());
            }

            @Override
            public void onDiscoveryStopped(String serviceType) {
                Log.d(TAG, "Discovery stopped for " + serviceType);
            }

            @Override
            public void onStartDiscoveryFailed(String serviceType, int errorCode) {
                Log.e(TAG, "Discovery start failed: " + errorCode);
                if (!completed && !cancelled) {
                    completed = true;
                    timeoutHandler.removeCallbacks(timeoutRunnable);
                    callback.onNotFound();
                }
            }

            @Override
            public void onStopDiscoveryFailed(String serviceType, int errorCode) {
                Log.e(TAG, "Discovery stop failed: " + errorCode);
            }
        };

        try {
            nsdManager.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, discoveryListener);
            timeoutHandler.postDelayed(timeoutRunnable, DISCOVERY_TIMEOUT_MS);
        } catch (Exception e) {
            Log.e(TAG, "Failed to start discovery", e);
            callback.onNotFound();
        }
    }

    private void resolveService(NsdServiceInfo serviceInfo, final StrategyCallback callback,
                                final Handler timeoutHandler, final Runnable timeoutRunnable) {
        try {
            nsdManager.resolveService(serviceInfo, new NsdManager.ResolveListener() {
                @Override
                public void onServiceResolved(NsdServiceInfo resolvedInfo) {
                    if (cancelled || completed) return;

                    String host = resolvedInfo.getHost().getHostAddress();
                    int port = resolvedInfo.getPort();
                    String candidateUrl = "http://" + host + ":" + port;

                    Log.d(TAG, "Resolved service at " + candidateUrl);

                    if (validator.validate(candidateUrl)) {
                        if (!completed && !cancelled) {
                            completed = true;
                            timeoutHandler.removeCallbacks(timeoutRunnable);
                            stopDiscovery();
                            callback.onFound(candidateUrl);
                        }
                    }
                }

                @Override
                public void onResolveFailed(NsdServiceInfo serviceInfo, int errorCode) {
                    Log.e(TAG, "Resolve failed: " + errorCode);
                }
            });
        } catch (Exception e) {
            Log.e(TAG, "Failed to resolve service", e);
        }
    }

    private void stopDiscovery() {
        try {
            if (nsdManager != null && discoveryListener != null) {
                nsdManager.stopServiceDiscovery(discoveryListener);
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to stop discovery", e);
        }
    }

    @Override
    public void cancel() {
        cancelled = true;
        stopDiscovery();
    }
}
