package com.kuranas.mobile.domain.repository;

import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.infra.http.ApiCallback;

public interface SearchRepository {

    void searchGlobal(String query, int limit, ApiCallback<SearchResult> callback);
}
