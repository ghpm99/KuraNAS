package com.kuranas.mobile.domain.model;

public final class DiscoveryResult {

    private final String serverUrl;
    private final String source;

    public DiscoveryResult(String serverUrl, String source) {
        if (serverUrl == null) {
            throw new IllegalArgumentException("serverUrl must not be null");
        }
        if (source == null) {
            throw new IllegalArgumentException("source must not be null");
        }
        this.serverUrl = serverUrl;
        this.source = source;
    }

    public String getServerUrl() {
        return serverUrl;
    }

    public String getSource() {
        return source;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        DiscoveryResult that = (DiscoveryResult) o;
        return serverUrl.equals(that.serverUrl) && source.equals(that.source);
    }

    @Override
    public int hashCode() {
        int result = serverUrl.hashCode();
        result = 31 * result + source.hashCode();
        return result;
    }

    @Override
    public String toString() {
        return "DiscoveryResult{serverUrl='" + serverUrl + "', source='" + source + "'}";
    }
}
