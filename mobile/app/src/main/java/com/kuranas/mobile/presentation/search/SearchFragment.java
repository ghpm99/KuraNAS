package com.kuranas.mobile.presentation.search;

import android.app.Activity;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.view.inputmethod.EditorInfo;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;

import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.repository.SearchRepository;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;
import com.kuranas.mobile.presentation.base.ViewState;
import com.kuranas.mobile.presentation.search.SearchResultItem;

public class SearchFragment extends BaseFragment {

    private static final int SEARCH_LIMIT = 50;

    private EditText searchInput;
    private Button btnSearch;
    private RecyclerView searchResults;

    private SearchAdapter adapter;
    private SearchRepository searchRepository;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_search, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        searchRepository = locator.getSearchRepository();

        searchInput = (EditText) root.findViewById(R.id.search_input);
        btnSearch = (Button) root.findViewById(R.id.btn_search);
        searchResults = (RecyclerView) root.findViewById(R.id.search_results);

        setupRecyclerView();

        btnSearch.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                performSearch();
            }
        });

        searchInput.setOnEditorActionListener(new TextView.OnEditorActionListener() {
            @Override
            public boolean onEditorAction(TextView v, int actionId, android.view.KeyEvent event) {
                if (actionId == EditorInfo.IME_ACTION_SEARCH) {
                    performSearch();
                    return true;
                }
                return false;
            }
        });

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                performSearch();
            }
        });

        return root;
    }

    private void setupRecyclerView() {
        searchResults.setLayoutManager(new LinearLayoutManager(getActivity()));

        adapter = new SearchAdapter(new SearchAdapter.OnResultClickListener() {
            @Override
            public void onResultClick(SearchResultItem item) {
                handleResultClick(item);
            }
        });

        searchResults.setAdapter(adapter);
    }

    private void performSearch() {
        String query = searchInput.getText().toString().trim();
        if (query.isEmpty()) {
            return;
        }

        setState(ViewState.LOADING);

        searchRepository.searchGlobal(query, SEARCH_LIMIT, new ApiCallback<SearchResult>() {
            @Override
            public void onSuccess(SearchResult result) {
                if (!isAdded()) {
                    return;
                }
                if (result.isEmpty()) {
                    setEmptyMessage(t("GLOBAL_SEARCH_EMPTY_TITLE"));
                    setState(ViewState.EMPTY);
                } else {
                    adapter.setResults(result);
                    setState(ViewState.CONTENT);
                }
            }

            @Override
            public void onError(AppError error) {
                if (!isAdded()) {
                    return;
                }
                setErrorMessage(error.getMessage());
                setState(ViewState.ERROR);
            }
        });
    }

    private void handleResultClick(SearchResultItem item) {
        Activity activity = getActivity();
        if (activity == null) {
            return;
        }
        if (!(activity instanceof SearchNavigationHost)) {
            return;
        }
        SearchNavigationHost host = (SearchNavigationHost) activity;

        SearchResultItem.Type type = item.getType();
        if (type == SearchResultItem.Type.FILE) {
            host.onSearchFileSelected(item.getFileItem());
        } else if (type == SearchResultItem.Type.FOLDER) {
            host.onSearchFolderSelected(item.getFileItem());
        } else if (type == SearchResultItem.Type.IMAGE) {
            host.onSearchImageSelected(item.getFileItem());
        } else if (type == SearchResultItem.Type.VIDEO) {
            host.onSearchVideoSelected(item.getVideoItem());
        }
    }

    public interface SearchNavigationHost {
        void onSearchFileSelected(FileItem file);
        void onSearchFolderSelected(FileItem folder);
        void onSearchImageSelected(FileItem image);
        void onSearchVideoSelected(VideoItem video);
    }
}
