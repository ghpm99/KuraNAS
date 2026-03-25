package com.kuranas.mobile.data.repository;

import com.kuranas.mobile.data.remote.api.SearchApi;
import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.domain.repository.SearchRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public final class SearchRepositoryImpl implements SearchRepository {

    private final SearchApi searchApi;

    public SearchRepositoryImpl(SearchApi searchApi) {
        this.searchApi = searchApi;
    }

    @Override
    public void searchGlobal(String query, int limit, ApiCallback<SearchResult> callback) {
        searchApi.searchGlobal(query, limit, callback);
    }
}
