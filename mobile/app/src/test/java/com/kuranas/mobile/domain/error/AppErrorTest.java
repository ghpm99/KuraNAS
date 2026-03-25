package com.kuranas.mobile.domain.error;

import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertNull;

public class AppErrorTest {

    @Test
    public void networkUnavailable_setsTypeAndMessage() {
        Exception cause = new Exception("no network");
        AppError error = AppError.networkUnavailable(cause);

        assertEquals(AppError.Type.NETWORK_UNAVAILABLE, error.getType());
        assertEquals("Network unavailable", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void timeout_setsTypeAndMessage() {
        Exception cause = new Exception("timed out");
        AppError error = AppError.timeout(cause);

        assertEquals(AppError.Type.TIMEOUT, error.getType());
        assertEquals("Request timed out", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void serverError_setsTypeAndIncludesStatusCode() {
        AppError error = AppError.serverError(503);

        assertEquals(AppError.Type.SERVER_ERROR, error.getType());
        assertEquals("Server error: 503", error.getMessage());
        assertNull(error.getCause());
    }

    @Test
    public void invalidPayload_setsTypeAndMessage() {
        Exception cause = new Exception("parse error");
        AppError error = AppError.invalidPayload(cause);

        assertEquals(AppError.Type.INVALID_PAYLOAD, error.getType());
        assertEquals("Invalid response data", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void unknown_setsTypeAndMessage() {
        Exception cause = new Exception("something went wrong");
        AppError error = AppError.unknown(cause);

        assertEquals(AppError.Type.UNKNOWN, error.getType());
        assertEquals("An unexpected error occurred", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void fromHttpResponse_500_returnsServerError() {
        Exception cause = new Exception("server failure");
        AppError error = AppError.fromHttpResponse(500, cause);

        assertEquals(AppError.Type.SERVER_ERROR, error.getType());
        assertEquals("Server error: 500", error.getMessage());
        assertNull(error.getCause());
    }

    @Test
    public void fromHttpResponse_502_returnsServerError() {
        AppError error = AppError.fromHttpResponse(502, null);

        assertEquals(AppError.Type.SERVER_ERROR, error.getType());
        assertEquals("Server error: 502", error.getMessage());
    }

    @Test
    public void fromHttpResponse_401_returnsUnauthorized() {
        Exception cause = new Exception("auth failed");
        AppError error = AppError.fromHttpResponse(401, cause);

        assertEquals(AppError.Type.UNAUTHORIZED, error.getType());
        assertEquals("Unauthorized", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void fromHttpResponse_403_returnsUnauthorized() {
        Exception cause = new Exception("forbidden");
        AppError error = AppError.fromHttpResponse(403, cause);

        assertEquals(AppError.Type.UNAUTHORIZED, error.getType());
        assertEquals("Unauthorized", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void fromHttpResponse_404_returnsServerErrorNotFound() {
        Exception cause = new Exception("not found");
        AppError error = AppError.fromHttpResponse(404, cause);

        assertEquals(AppError.Type.SERVER_ERROR, error.getType());
        assertEquals("Not found", error.getMessage());
        assertNull(error.getCause());
    }

    @Test
    public void fromHttpResponse_otherStatusCode_returnsServerErrorWithHttp() {
        Exception cause = new Exception("bad request");
        AppError error = AppError.fromHttpResponse(400, cause);

        assertEquals(AppError.Type.SERVER_ERROR, error.getType());
        assertEquals("HTTP 400", error.getMessage());
        assertEquals(cause, error.getCause());
    }

    @Test
    public void constructor_withoutCause_causeIsNull() {
        AppError error = new AppError(AppError.Type.UNKNOWN, "test");

        assertEquals(AppError.Type.UNKNOWN, error.getType());
        assertEquals("test", error.getMessage());
        assertNull(error.getCause());
    }

    @Test
    public void constructor_withCause_causeIsSet() {
        Exception cause = new Exception("root cause");
        AppError error = new AppError(AppError.Type.TIMEOUT, "msg", cause);

        assertNotNull(error.getCause());
        assertEquals(cause, error.getCause());
    }
}
