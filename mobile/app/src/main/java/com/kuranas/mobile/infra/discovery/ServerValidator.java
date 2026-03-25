package com.kuranas.mobile.infra.discovery;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;

public class ServerValidator {

    private static final int VALIDATION_TIMEOUT_MS = 3000;
    private static final String HEALTH_PATH = "/api/v1/health";

    public boolean validate(String candidateUrl) {
        HttpURLConnection connection = null;
        try {
            String url = candidateUrl + HEALTH_PATH;
            connection = (HttpURLConnection) new URL(url).openConnection();
            connection.setConnectTimeout(VALIDATION_TIMEOUT_MS);
            connection.setReadTimeout(VALIDATION_TIMEOUT_MS);
            connection.setRequestMethod("GET");

            int responseCode = connection.getResponseCode();
            if (responseCode != 200) {
                return false;
            }

            BufferedReader reader = new BufferedReader(
                    new InputStreamReader(connection.getInputStream()));
            StringBuilder body = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                body.append(line);
            }
            reader.close();

            return body.toString().contains("kuranas");
        } catch (Exception e) {
            return false;
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }
}
