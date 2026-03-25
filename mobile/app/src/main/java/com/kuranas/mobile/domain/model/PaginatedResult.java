package com.kuranas.mobile.domain.model;

import java.util.List;

public final class PaginatedResult<T> {

    private final List<T> items;
    private final Pagination pagination;

    public PaginatedResult(List<T> items, Pagination pagination) {
        this.items = items;
        this.pagination = pagination;
    }

    public List<T> getItems() {
        return items;
    }

    public Pagination getPagination() {
        return pagination;
    }

    public boolean hasNext() {
        return pagination != null && pagination.hasNext();
    }
}
