package com.kuranas.mobile.domain.model;

/**
 * A single notification shown on the kiosk panel. Lean by design: only the
 * fields the wall panel renders (type drives the colour marker, title/message
 * the two text lines, the timestamp the time on the side).
 */
public final class NotificationItem {

    private final String type;
    private final String title;
    private final String message;
    private final String createdAt;

    public NotificationItem(String type, String title, String message, String createdAt) {
        this.type = type;
        this.title = title;
        this.message = message;
        this.createdAt = createdAt;
    }

    public String getType() {
        return type;
    }

    public String getTitle() {
        return title;
    }

    public String getMessage() {
        return message;
    }

    public String getCreatedAt() {
        return createdAt;
    }
}
