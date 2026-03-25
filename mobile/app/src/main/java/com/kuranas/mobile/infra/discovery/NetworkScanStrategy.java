package com.kuranas.mobile.infra.discovery;

import android.content.Context;
import android.net.wifi.WifiInfo;
import android.net.wifi.WifiManager;
import android.util.Log;

import java.net.InetSocketAddress;
import java.net.Socket;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

public final class NetworkScanStrategy implements DiscoveryStrategy {

    private static final String TAG = "NetworkScan";
    private static final int TARGET_PORT = 8000;
    private static final int SOCKET_TIMEOUT_MS = 300;
    private static final int OVERALL_TIMEOUT_S = 15;
    private static final int THREAD_POOL_SIZE = 20;

    private final Context context;
    private final ServerValidator validator;
    private volatile boolean cancelled;
    private ExecutorService executor;

    public NetworkScanStrategy(Context context, ServerValidator validator) {
        this.context = context.getApplicationContext();
        this.validator = validator;
    }

    @Override
    public String name() {
        return "scan";
    }

    @Override
    public void discover(final StrategyCallback callback) {
        cancelled = false;

        String subnet = getSubnet();
        if (subnet == null) {
            Log.e(TAG, "Could not determine device subnet");
            callback.onNotFound();
            return;
        }

        Log.d(TAG, "Scanning subnet: " + subnet + ".* on port " + TARGET_PORT);

        executor = Executors.newFixedThreadPool(THREAD_POOL_SIZE);
        final AtomicReference<String> foundUrl = new AtomicReference<String>(null);
        final CountDownLatch latch = new CountDownLatch(254);

        for (int i = 1; i <= 254; i++) {
            final String ip = subnet + "." + i;

            executor.submit(new Runnable() {
                @Override
                public void run() {
                    try {
                        if (cancelled || foundUrl.get() != null) return;

                        if (isPortOpen(ip, TARGET_PORT)) {
                            String candidateUrl = "http://" + ip + ":" + TARGET_PORT;
                            Log.d(TAG, "Port open at " + candidateUrl + ", validating...");

                            if (validator.validate(candidateUrl)) {
                                foundUrl.compareAndSet(null, candidateUrl);
                            }
                        }
                    } finally {
                        latch.countDown();
                    }
                }
            });
        }

        try {
            latch.await(OVERALL_TIMEOUT_S, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            Log.d(TAG, "Scan interrupted");
        }

        executor.shutdownNow();
        executor = null;

        String result = foundUrl.get();
        if (result != null && !cancelled) {
            callback.onFound(result);
        } else {
            callback.onNotFound();
        }
    }

    private boolean isPortOpen(String ip, int port) {
        Socket socket = null;
        try {
            socket = new Socket();
            socket.connect(new InetSocketAddress(ip, port), SOCKET_TIMEOUT_MS);
            return true;
        } catch (Exception e) {
            return false;
        } finally {
            if (socket != null) {
                try {
                    socket.close();
                } catch (Exception ignored) {
                }
            }
        }
    }

    String getSubnet() {
        try {
            WifiManager wifiManager = (WifiManager) context.getSystemService(Context.WIFI_SERVICE);
            if (wifiManager == null) return null;

            WifiInfo wifiInfo = wifiManager.getConnectionInfo();
            int ip = wifiInfo.getIpAddress();

            if (ip == 0) return null;

            return (ip & 0xFF) + "." +
                    ((ip >> 8) & 0xFF) + "." +
                    ((ip >> 16) & 0xFF);
        } catch (Exception e) {
            Log.e(TAG, "Failed to get subnet", e);
            return null;
        }
    }

    @Override
    public void cancel() {
        cancelled = true;
        ExecutorService exec = executor;
        if (exec != null) {
            exec.shutdownNow();
        }
    }
}
