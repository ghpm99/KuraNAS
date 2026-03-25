package com.kuranas.mobile.presentation.files;

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
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;
import com.kuranas.mobile.presentation.base.ViewState;

import java.util.ArrayList;
import java.util.List;

public class FilesFragment extends BaseFragment {

    private static final int PAGE_SIZE = 30;

    private SwipeRefreshLayout swipeRefresh;
    private RecyclerView filesList;
    private TextView currentPathView;

    private FilesAdapter adapter;
    private FileRepository fileRepository;
    private TranslationManager translations;
    private BitmapCache bitmapCache;
    private String baseUrl;

    private String currentPath = "";
    private int currentPage = 1;
    private boolean hasNextPage = false;
    private boolean isLoadingMore = false;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_files, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        fileRepository = locator.getFileRepository();
        translations = locator.getTranslationManager();
        bitmapCache = locator.getBitmapCache();
        baseUrl = locator.getHttpClient().getBaseUrl();

        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        filesList = (RecyclerView) root.findViewById(R.id.files_list);
        currentPathView = (TextView) root.findViewById(R.id.current_path);

        setupRecyclerView();

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                loadFiles(true);
            }
        });

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadFiles(true);
            }
        });

        loadFiles(true);

        return root;
    }

    private void setupRecyclerView() {
        final LinearLayoutManager layoutManager = new LinearLayoutManager(getActivity());
        filesList.setLayoutManager(layoutManager);

        adapter = new FilesAdapter(new ArrayList<FileItem>(), new FilesAdapter.OnItemClickListener() {
            @Override
            public void onItemClick(FileItem item) {
                handleItemClick(item);
            }
        }, translations, bitmapCache, baseUrl);

        filesList.setAdapter(adapter);

        filesList.addOnScrollListener(new RecyclerView.OnScrollListener() {
            @Override
            public void onScrolled(RecyclerView recyclerView, int dx, int dy) {
                super.onScrolled(recyclerView, dx, dy);
                if (isLoadingMore || !hasNextPage) {
                    return;
                }
                int totalItemCount = layoutManager.getItemCount();
                int lastVisibleItem = layoutManager.findLastVisibleItemPosition();
                if (lastVisibleItem >= totalItemCount - 5) {
                    loadMoreFiles();
                }
            }
        });
    }

    private void loadFiles(boolean reset) {
        if (reset) {
            currentPage = 1;
        }
        if (currentPage == 1) {
            setState(ViewState.LOADING);
        }
        isLoadingMore = currentPage > 1;

        updateBreadcrumb();

        fileRepository.getFilesByPath(currentPath, currentPage, PAGE_SIZE,
                new ApiCallback<PaginatedResult<FileItem>>() {
                    @Override
                    public void onSuccess(PaginatedResult<FileItem> result) {
                        if (!isAdded()) {
                            return;
                        }
                        swipeRefresh.setRefreshing(false);
                        isLoadingMore = false;
                        hasNextPage = result.hasNext();

                        List<FileItem> items = result.getItems();
                        if (currentPage == 1) {
                            if (items != null && !items.isEmpty()) {
                                adapter.updateItems(items);
                                setState(ViewState.CONTENT);
                            } else {
                                setEmptyMessage(t("files.empty_folder"));
                                setState(ViewState.EMPTY);
                            }
                        } else {
                            if (items != null && !items.isEmpty()) {
                                adapter.addItems(items);
                            }
                        }
                    }

                    @Override
                    public void onError(AppError error) {
                        if (!isAdded()) {
                            return;
                        }
                        swipeRefresh.setRefreshing(false);
                        isLoadingMore = false;
                        if (currentPage == 1) {
                            setErrorMessage(error.getMessage());
                            setState(ViewState.ERROR);
                        }
                    }
                });
    }

    private void loadMoreFiles() {
        currentPage++;
        loadFiles(false);
    }

    private void handleItemClick(FileItem item) {
        if (item.isDirectory()) {
            navigateTo(item.getPath());
        } else if (item.isImage()) {
            Activity activity = getActivity();
            if (activity instanceof FileNavigationHost) {
                ((FileNavigationHost) activity).openImageViewer(item);
            }
        } else if (item.isAudio()) {
            Activity activity = getActivity();
            if (activity instanceof FileNavigationHost) {
                ((FileNavigationHost) activity).openMusicPlayer(item);
            }
        } else if (item.isVideo()) {
            Activity activity = getActivity();
            if (activity instanceof FileNavigationHost) {
                ((FileNavigationHost) activity).openVideoPlayer(item);
            }
        } else {
            Activity activity = getActivity();
            if (activity instanceof FileNavigationHost) {
                ((FileNavigationHost) activity).openFile(item);
            }
        }
    }

    public void navigateTo(String path) {
        currentPath = path != null ? path : "";
        loadFiles(true);
    }

    public boolean handleBackNavigation() {
        if (currentPath == null || currentPath.isEmpty()) {
            return false;
        }
        int lastSlash = currentPath.lastIndexOf('/');
        if (lastSlash > 0) {
            currentPath = currentPath.substring(0, lastSlash);
        } else {
            currentPath = "";
        }
        loadFiles(true);
        return true;
    }

    private void updateBreadcrumb() {
        if (currentPathView == null) {
            return;
        }
        if (currentPath == null || currentPath.isEmpty()) {
            currentPathView.setText("/");
        } else {
            currentPathView.setText(currentPath);
        }
    }

    public interface FileNavigationHost {
        void openImageViewer(FileItem item);
        void openMusicPlayer(FileItem item);
        void openVideoPlayer(FileItem item);
        void openFile(FileItem item);
    }
}
