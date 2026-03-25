package com.kuranas.mobile.infra.http;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

public final class HttpResponse {

    private final int statusCode;
    private final String body;
    private final Exception error;

    private HttpResponse(int statusCode, String body, Exception error) {
        this.statusCode = statusCode;
        this.body = body;
        this.error = error;
    }

    public static HttpResponse success(int statusCode, String body) {
        return new HttpResponse(statusCode, body, null);
    }

    public static HttpResponse failure(Exception error) {
        return new HttpResponse(-1, null, error);
    }

    public boolean isSuccessful() {
        return statusCode >= 200 && statusCode < 300 && error == null;
    }

    public int getStatusCode() {
        return statusCode;
    }

    public String getBody() {
        return body;
    }

    public Exception getError() {
        return error;
    }

    public JSONObject toJsonObject() throws JSONException {
        if (body == null) {
            throw new JSONException("Response body is null");
        }
        return new JSONObject(body);
    }

    public JSONArray toJsonArray() throws JSONException {
        if (body == null) {
            throw new JSONException("Response body is null");
        }
        return new JSONArray(body);
    }
}
