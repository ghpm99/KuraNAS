package com.kuranas.mobile.domain.model;

public final class Pagination {

    private final int page;
    private final int pageSize;
    private final boolean hasNext;
    private final boolean hasPrev;

    public Pagination(int page, int pageSize, boolean hasNext, boolean hasPrev) {
        this.page = page;
        this.pageSize = pageSize;
        this.hasNext = hasNext;
        this.hasPrev = hasPrev;
    }

    public int getPage() {
        return page;
    }

    public int getPageSize() {
        return pageSize;
    }

    public boolean hasNext() {
        return hasNext;
    }

    public boolean hasPrev() {
        return hasPrev;
    }
}
