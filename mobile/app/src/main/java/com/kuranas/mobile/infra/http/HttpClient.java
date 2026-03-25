package com.kuranas.mobile.infra.http;

import android.os.Handler;
import android.os.Looper;

import com.kuranas.mobile.infra.logging.AppLogger;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public final class HttpClient {

    private static final String LOG_TAG = "HttpClient";
    private static final int CONNECT_TIMEOUT_MS = 10000;
    private static final int READ_TIMEOUT_MS = 15000;
    private static final int THREAD_POOL_SIZE = 4;

    private final String baseUrl;
    private final ExecutorService executor;
    private final Handler mainHandler;

    public HttpClient(String baseUrl) {
        this.baseUrl = baseUrl;
        this.executor = Executors.newFixedThreadPool(THREAD_POOL_SIZE);
        this.mainHandler = new Handler(Looper.getMainLooper());
    }

    public void get(String path, final Callback callback) {
        request("GET", path, null, callback);
    }

    public void post(String path, String jsonBody, final Callback callback) {
        request("POST", path, jsonBody, callback);
    }

    public void put(String path, String jsonBody, final Callback callback) {
        request("PUT", path, jsonBody, callback);
    }

    public HttpResponse getSync(String path) {
        return requestSync("GET", path, null);
    }

    public HttpResponse postSync(String path, String jsonBody) {
        return requestSync("POST", path, jsonBody);
    }

    public HttpResponse putSync(String path, String jsonBody) {
        return requestSync("PUT", path, jsonBody);
    }

    public String getBaseUrl() {
        return baseUrl;
    }

    public void shutdown() {
        executor.shutdown();
    }

    private void request(final String method, final String path, final String jsonBody, final Callback callback) {
        executor.execute(new Runnable() {
            @Override
            public void run() {
                final HttpResponse response = requestSync(method, path, jsonBody);
                mainHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        callback.onResponse(response);
                    }
                });
            }
        });
    }

    private HttpResponse requestSync(String method, String path, String jsonBody) {
        HttpURLConnection connection = null;
        try {
            String fullUrl = baseUrl + path;
            AppLogger.d(LOG_TAG, method + " " + fullUrl);

            URL url = new URL(fullUrl);
            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod(method);
            connection.setConnectTimeout(CONNECT_TIMEOUT_MS);
            connection.setReadTimeout(READ_TIMEOUT_MS);
            connection.setRequestProperty("Accept", "application/json");

            if (jsonBody != null) {
                connection.setRequestProperty("Content-Type", "application/json; charset=UTF-8");
                connection.setDoOutput(true);
                OutputStream os = connection.getOutputStream();
                try {
                    os.write(jsonBody.getBytes("UTF-8"));
                    os.flush();
                } finally {
                    os.close();
                }
            }

            int statusCode = connection.getResponseCode();
            InputStream inputStream;
            if (statusCode >= 200 && statusCode < 300) {
                inputStream = connection.getInputStream();
            } else {
                inputStream = connection.getErrorStream();
            }

            String body = readStream(inputStream);
            return HttpResponse.success(statusCode, body);

        } catch (IOException e) {
            AppLogger.e(LOG_TAG, "Request failed: " + path, e);
            return HttpResponse.failure(e);
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    private String readStream(InputStream inputStream) throws IOException {
        if (inputStream == null) {
            return "";
        }
        BufferedReader reader = new BufferedReader(new InputStreamReader(inputStream, "UTF-8"));
        try {
            StringBuilder sb = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line);
            }
            return sb.toString();
        } finally {
            reader.close();
        }
    }

    public interface Callback {
        void onResponse(HttpResponse response);
    }
}
