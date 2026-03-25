package com.kuranas.mobile.presentation.home;

import android.app.Activity;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import androidx.swiperefreshlayout.widget.SwipeRefreshLayout;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.repository.FileRepository;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;
import com.kuranas.mobile.presentation.base.ViewState;

import java.util.ArrayList;
import java.util.List;

public class HomeFragment extends BaseFragment {

    private SwipeRefreshLayout swipeRefresh;
    private RecyclerView recentFilesList;
    private RecyclerView starredList;
    private TextView sectionRecentTitle;
    private TextView sectionStarredTitle;

    private HomeSectionAdapter recentAdapter;
    private HomeSectionAdapter starredAdapter;

    private FileRepository fileRepository;
    private BitmapCache bitmapCache;
    private String baseUrl;

    private boolean recentLoaded;
    private boolean starredLoaded;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_home, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        fileRepository = locator.getFileRepository();
        bitmapCache = locator.getBitmapCache();
        baseUrl = locator.getHttpClient().getBaseUrl();

        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        recentFilesList = (RecyclerView) root.findViewById(R.id.recent_files_list);
        starredList = (RecyclerView) root.findViewById(R.id.starred_list);
        sectionRecentTitle = (TextView) root.findViewById(R.id.section_recent_title);
        sectionStarredTitle = (TextView) root.findViewById(R.id.section_starred_title);

        sectionRecentTitle.setText(t("home.recent_files"));
        sectionStarredTitle.setText(t("home.starred"));

        setupRecyclerViews();

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                loadData();
            }
        });

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadData();
            }
        });

        loadData();

        return root;
    }

    private void setupRecyclerViews() {
        recentFilesList.setLayoutManager(
                new LinearLayoutManager(getActivity(), LinearLayoutManager.HORIZONTAL, false));
        starredList.setLayoutManager(
                new LinearLayoutManager(getActivity(), LinearLayoutManager.HORIZONTAL, false));

        HomeSectionAdapter.OnItemClickListener clickListener = new HomeSectionAdapter.OnItemClickListener() {
            @Override
            public void onItemClick(FileItem item) {
                handleItemClick(item);
            }
        };

        recentAdapter = new HomeSectionAdapter(
                new ArrayList<FileItem>(), clickListener, bitmapCache, baseUrl);
        starredAdapter = new HomeSectionAdapter(
                new ArrayList<FileItem>(), clickListener, bitmapCache, baseUrl);

        recentFilesList.setAdapter(recentAdapter);
        starredList.setAdapter(starredAdapter);
    }

    private void loadData() {
        recentLoaded = false;
        starredLoaded = false;
        setState(ViewState.LOADING);

        fileRepository.getFileTree(1, 10, new ApiCallback<PaginatedResult<FileItem>>() {
            @Override
            public void onSuccess(PaginatedResult<FileItem> result) {
                if (!isAdded()) {
                    return;
                }
                List<FileItem> items = result.getItems();
                if (items != null && !items.isEmpty()) {
                    recentAdapter.updateItems(items);
                    sectionRecentTitle.setVisibility(View.VISIBLE);
                    recentFilesList.setVisibility(View.VISIBLE);
                } else {
                    sectionRecentTitle.setVisibility(View.GONE);
                    recentFilesList.setVisibility(View.GONE);
                }
                recentLoaded = true;
                checkLoadComplete();
            }

            @Override
            public void onError(AppError error) {
                if (!isAdded()) {
                    return;
                }
                recentLoaded = true;
                sectionRecentTitle.setVisibility(View.GONE);
                recentFilesList.setVisibility(View.GONE);
                checkLoadComplete();
            }
        });

        fileRepository.getStarredFiles(1, 10, new ApiCallback<PaginatedResult<FileItem>>() {
            @Override
            public void onSuccess(PaginatedResult<FileItem> result) {
                if (!isAdded()) {
                    return;
                }
                List<FileItem> items = result.getItems();
                if (items != null && !items.isEmpty()) {
                    starredAdapter.updateItems(items);
                    sectionStarredTitle.setVisibility(View.VISIBLE);
                    starredList.setVisibility(View.VISIBLE);
                } else {
                    sectionStarredTitle.setVisibility(View.GONE);
                    starredList.setVisibility(View.GONE);
                }
                starredLoaded = true;
                checkLoadComplete();
            }

            @Override
            public void onError(AppError error) {
                if (!isAdded()) {
                    return;
                }
                starredLoaded = true;
                sectionStarredTitle.setVisibility(View.GONE);
                starredList.setVisibility(View.GONE);
                checkLoadComplete();
            }
        });
    }

    private void checkLoadComplete() {
        if (!recentLoaded || !starredLoaded) {
            return;
        }
        swipeRefresh.setRefreshing(false);
        boolean hasRecent = recentAdapter.getItemCount() > 0;
        boolean hasStarred = starredAdapter.getItemCount() > 0;
        if (hasRecent || hasStarred) {
            setState(ViewState.CONTENT);
        } else {
            setEmptyMessage(t("home.empty"));
            setState(ViewState.EMPTY);
        }
    }

    private void handleItemClick(FileItem item) {
        Activity activity = getActivity();
        if (activity == null) {
            return;
        }
        if (activity instanceof NavigationHost) {
            ((NavigationHost) activity).onFileItemSelected(item);
        }
    }

    public interface NavigationHost {
        void onFileItemSelected(FileItem item);
    }
}
