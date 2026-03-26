package com.kuranas.mobile.presentation.video;

import android.content.Intent;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;

import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import androidx.swiperefreshlayout.widget.SwipeRefreshLayout;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.repository.VideoRepository;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;
import com.kuranas.mobile.presentation.base.ViewState;

import java.util.ArrayList;
import java.util.List;

public class VideoFragment extends BaseFragment {

    private static final int PAGE_SIZE = 30;

    private SwipeRefreshLayout swipeRefresh;
    private RecyclerView videoList;

    private VideoAdapter adapter;
    private VideoRepository videoRepository;
    private BitmapCache bitmapCache;
    private String baseUrl;

    private int currentPage = 1;
    private boolean hasNextPage = false;
    private boolean isLoadingMore = false;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_video, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        videoRepository = locator.getVideoRepository();
        bitmapCache = locator.getBitmapCache();
        baseUrl = locator.getHttpClient().getBaseUrl();

        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        videoList = (RecyclerView) root.findViewById(R.id.video_list);

        setupRecyclerView();

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                loadVideos(true);
            }
        });

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadVideos(true);
            }
        });

        loadVideos(true);

        return root;
    }

    private void setupRecyclerView() {
        final LinearLayoutManager layoutManager = new LinearLayoutManager(getActivity());
        videoList.setLayoutManager(layoutManager);

        adapter = new VideoAdapter(new ArrayList<VideoItem>(), new VideoAdapter.OnVideoClickListener() {
            @Override
            public void onVideoClick(VideoItem video) {
                openVideoPlayer(video);
            }
        }, bitmapCache, baseUrl);

        videoList.setAdapter(adapter);

        videoList.addOnScrollListener(new RecyclerView.OnScrollListener() {
            @Override
            public void onScrolled(RecyclerView recyclerView, int dx, int dy) {
                super.onScrolled(recyclerView, dx, dy);
                if (isLoadingMore || !hasNextPage) {
                    return;
                }
                int totalItemCount = layoutManager.getItemCount();
                int lastVisibleItem = layoutManager.findLastVisibleItemPosition();
                if (lastVisibleItem >= totalItemCount - 5) {
                    loadMoreVideos();
                }
            }
        });
    }

    private void loadVideos(boolean reset) {
        if (reset) {
            currentPage = 1;
        }
        if (currentPage == 1) {
            setState(ViewState.LOADING);
        }
        isLoadingMore = currentPage > 1;

        videoRepository.getLibraryVideos(currentPage, PAGE_SIZE,
                new ApiCallback<PaginatedResult<VideoItem>>() {
                    @Override
                    public void onSuccess(PaginatedResult<VideoItem> result) {
                        if (!isUiReady()) {
                            return;
                        }
                        swipeRefresh.setRefreshing(false);
                        isLoadingMore = false;
                        hasNextPage = result.hasNext();

                        List<VideoItem> items = result.getItems();
                        if (currentPage == 1) {
                            if (items != null && !items.isEmpty()) {
                                adapter.updateItems(items);
                                setState(ViewState.CONTENT);
                            } else {
                                setEmptyMessage(t("HOME_VIDEO_EMPTY"));
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

    private void loadMoreVideos() {
        currentPage++;
        loadVideos(false);
    }

    private void openVideoPlayer(VideoItem video) {
        Intent intent = new Intent(getActivity(), VideoPlayerActivity.class);
        intent.putExtra(VideoPlayerActivity.EXTRA_VIDEO_ID, video.getId());
        intent.putExtra(VideoPlayerActivity.EXTRA_VIDEO_NAME, video.getDisplayName());
        intent.putExtra(VideoPlayerActivity.EXTRA_STREAM_URL,
                baseUrl + "/api/v1/files/video-stream/" + video.getId());
        startActivity(intent);
    }
}
