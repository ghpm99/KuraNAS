package com.kuranas.mobile.domain.model;

/**
 * A single e-mail shown on the kiosk panel. Carries no body — only the metadata
 * the panel renders. {@code verdict}/{@code importance}/{@code summary} are only
 * present once the message has been analyzed (task 16); they may be empty. The
 * panel never renders HTML: every field arrives as plain text from the backend.
 */
public final class EmailItem {

    private final String senderName;
    private final String senderAddress;
    private final String subject;
    private final String snippet;
    private final String summary;
    private final String importance;
    private final String verdict;
    private final String receivedAt;

    public EmailItem(String senderName, String senderAddress, String subject, String snippet,
                     String summary, String importance, String verdict, String receivedAt) {
        this.senderName = senderName;
        this.senderAddress = senderAddress;
        this.subject = subject;
        this.snippet = snippet;
        this.summary = summary;
        this.importance = importance;
        this.verdict = verdict;
        this.receivedAt = receivedAt;
    }

    public String getSenderName() {
        return senderName;
    }

    public String getSenderAddress() {
        return senderAddress;
    }

    public String getSubject() {
        return subject;
    }

    public String getSnippet() {
        return snippet;
    }

    public String getSummary() {
        return summary;
    }

    public String getImportance() {
        return importance;
    }

    public String getVerdict() {
        return verdict;
    }

    public String getReceivedAt() {
        return receivedAt;
    }

    /** True when the AI flagged this message as malicious or suspicious. */
    public boolean isFlagged() {
        return "malicious".equals(verdict) || "suspicious".equals(verdict);
    }

    /** True when the AI marked this message as high importance. */
    public boolean isHighImportance() {
        return "high".equals(importance);
    }
}
