package com.kuranas.mobile.domain.error;

public final class AppError {

    private final Type type;
    private final String message;
    private final Exception cause;

    public AppError(Type type, String message) {
        this(type, message, null);
    }

    public AppError(Type type, String message, Exception cause) {
        this.type = type;
        this.message = message;
        this.cause = cause;
    }

    public Type getType() {
        return type;
    }

    public String getMessage() {
        return message;
    }

    public Exception getCause() {
        return cause;
    }

    public static AppError networkUnavailable(Exception cause) {
        return new AppError(Type.NETWORK_UNAVAILABLE, "Network unavailable", cause);
    }

    public static AppError timeout(Exception cause) {
        return new AppError(Type.TIMEOUT, "Request timed out", cause);
    }

    public static AppError serverError(int statusCode) {
        return new AppError(Type.SERVER_ERROR, "Server error: " + statusCode);
    }

    public static AppError invalidPayload(Exception cause) {
        return new AppError(Type.INVALID_PAYLOAD, "Invalid response data", cause);
    }

    public static AppError unknown(Exception cause) {
        return new AppError(Type.UNKNOWN, "An unexpected error occurred", cause);
    }

    public static AppError fromHttpResponse(int statusCode, Exception cause) {
        if (statusCode >= 500) {
            return serverError(statusCode);
        }
        if (statusCode == 401 || statusCode == 403) {
            return new AppError(Type.UNAUTHORIZED, "Unauthorized", cause);
        }
        if (statusCode == 404) {
            return new AppError(Type.SERVER_ERROR, "Not found");
        }
        return new AppError(Type.SERVER_ERROR, "HTTP " + statusCode, cause);
    }

    public enum Type {
        NETWORK_UNAVAILABLE,
        TIMEOUT,
        UNAUTHORIZED,
        SERVER_ERROR,
        INVALID_PAYLOAD,
        UNKNOWN
    }
}
