package com.kuranas.mobile.presentation.images;

import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;

import androidx.fragment.app.FragmentActivity;
import androidx.recyclerview.widget.GridLayoutManager;
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

public class ImagesFragment extends BaseFragment {

    private static final int PAGE_SIZE = 30;
    private static final int GRID_COLUMNS = 3;

    private SwipeRefreshLayout swipeRefresh;
    private RecyclerView imagesGrid;

    private ImagesAdapter adapter;
    private FileRepository fileRepository;
    private BitmapCache bitmapCache;
    private String baseUrl;

    private final List<FileItem> allImages = new ArrayList<FileItem>();
    private int currentPage = 1;
    private boolean hasNextPage = false;
    private boolean isLoadingMore = false;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_images, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        fileRepository = locator.getFileRepository();
        bitmapCache = locator.getBitmapCache();
        baseUrl = locator.getHttpClient().getBaseUrl();

        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        imagesGrid = (RecyclerView) root.findViewById(R.id.images_grid);

        setupRecyclerView();

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                loadImages(true);
            }
        });

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadImages(true);
            }
        });

        loadImages(true);

        return root;
    }

    private void setupRecyclerView() {
        final GridLayoutManager layoutManager = new GridLayoutManager(getActivity(), GRID_COLUMNS);
        imagesGrid.setLayoutManager(layoutManager);

        adapter = new ImagesAdapter(new ArrayList<FileItem>(), new ImagesAdapter.OnImageClickListener() {
            @Override
            public void onImageClick(int position) {
                openImageViewer(position);
            }
        }, bitmapCache, baseUrl);

        imagesGrid.setAdapter(adapter);

        imagesGrid.addOnScrollListener(new RecyclerView.OnScrollListener() {
            @Override
            public void onScrolled(RecyclerView recyclerView, int dx, int dy) {
                super.onScrolled(recyclerView, dx, dy);
                if (isLoadingMore || !hasNextPage) {
                    return;
                }
                int totalItemCount = layoutManager.getItemCount();
                int lastVisibleItem = layoutManager.findLastVisibleItemPosition();
                if (lastVisibleItem >= totalItemCount - 5) {
                    loadMoreImages();
                }
            }
        });
    }

    private void loadImages(boolean reset) {
        if (reset) {
            currentPage = 1;
            allImages.clear();
        }
        if (currentPage == 1) {
            setState(ViewState.LOADING);
        }
        isLoadingMore = currentPage > 1;

        fileRepository.getImages(currentPage, PAGE_SIZE, "date",
                new ApiCallback<PaginatedResult<FileItem>>() {
                    @Override
                    public void onSuccess(PaginatedResult<FileItem> result) {
                        if (!isUiReady()) {
                            return;
                        }
                        swipeRefresh.setRefreshing(false);
                        isLoadingMore = false;
                        hasNextPage = result.hasNext();

                        List<FileItem> items = result.getItems();
                        if (currentPage == 1) {
                            if (items != null && !items.isEmpty()) {
                                allImages.addAll(items);
                                adapter.updateItems(items);
                                setState(ViewState.CONTENT);
                            } else {
                                setEmptyMessage(t("IMAGES_EMPTY_TITLE"));
                                setState(ViewState.EMPTY);
                            }
                        } else {
                            if (items != null && !items.isEmpty()) {
                                allImages.addAll(items);
                                adapter.addItems(items);
                            }
                        }
                    }

                    @Override
                    public void onError(AppError error) {
                        if (!isUiReady()) {
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

    private boolean isUiReady() {
        return isAdded() && getView() != null && swipeRefresh != null && adapter != null;
    }

    private void loadMoreImages() {
        currentPage++;
        loadImages(false);
    }

    private void openImageViewer(int position) {
        ArrayList<Integer> imageIds = new ArrayList<Integer>();
        for (int i = 0; i < allImages.size(); i++) {
            imageIds.add(Integer.valueOf(allImages.get(i).getId()));
        }
        ImageViewerFragment viewer = ImageViewerFragment.newInstance(imageIds, position);
        FragmentActivity activity = getActivity();
        if (activity != null) {
            activity.getSupportFragmentManager()
                    .beginTransaction()
                    .replace(R.id.content_frame, viewer)
                    .addToBackStack(null)
                    .commit();
        }
    }
}
