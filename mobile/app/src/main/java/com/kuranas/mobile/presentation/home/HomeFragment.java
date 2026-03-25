package com.kuranas.mobile.presentation.home;

import android.app.Activity;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
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

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.List;
import java.util.Locale;

public class HomeFragment extends BaseFragment {

    private static final long CLOCK_UPDATE_INTERVAL_MS = 15000;

    private TextView clockText;
    private TextView dateText;
    private SwipeRefreshLayout swipeRefresh;
    private View sectionRecent;
    private View sectionStarred;
    private RecyclerView recentFilesList;
    private RecyclerView starredList;
    private TextView sectionRecentTitle;
    private TextView sectionStarredTitle;

    private HomeSectionAdapter recentAdapter;
    private HomeSectionAdapter starredAdapter;

    private FileRepository fileRepository;
    private BitmapCache bitmapCache;
    private String baseUrl;

    private final Handler clockHandler = new Handler(Looper.getMainLooper());
    private final Runnable clockRunnable = new Runnable() {
        @Override
        public void run() {
            updateClock();
            clockHandler.postDelayed(this, CLOCK_UPDATE_INTERVAL_MS);
        }
    };

    private boolean recentLoaded;
    private boolean starredLoaded;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_home, container, false);

        clockText = (TextView) root.findViewById(R.id.clock_text);
        dateText = (TextView) root.findViewById(R.id.date_text);
        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        sectionRecent = root.findViewById(R.id.section_recent);
        sectionStarred = root.findViewById(R.id.section_starred);
        recentFilesList = (RecyclerView) root.findViewById(R.id.recent_files_list);
        starredList = (RecyclerView) root.findViewById(R.id.starred_list);
        sectionRecentTitle = (TextView) root.findViewById(R.id.section_recent_title);
        sectionStarredTitle = (TextView) root.findViewById(R.id.section_starred_title);

        sectionRecentTitle.setText(t("RECENT_FILES"));
        sectionStarredTitle.setText(t("STARRED_FILES"));

        ServiceLocator locator = ServiceLocator.getInstance();
        fileRepository = locator.getFileRepository();
        bitmapCache = locator.getBitmapCache();
        baseUrl = locator.getHttpClient().getBaseUrl();

        setupRecyclerViews();
        updateClock();

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadData();
            }
        });

        loadData();

        return root;
    }

    @Override
    public void onResume() {
        super.onResume();
        updateClock();
        clockHandler.postDelayed(clockRunnable, CLOCK_UPDATE_INTERVAL_MS);
    }

    @Override
    public void onPause() {
        super.onPause();
        clockHandler.removeCallbacks(clockRunnable);
    }

    private void updateClock() {
        Calendar now = Calendar.getInstance();
        SimpleDateFormat timeFormat = new SimpleDateFormat("HH:mm", Locale.getDefault());
        SimpleDateFormat dateFormat = new SimpleDateFormat("EEEE, d 'de' MMMM", new Locale("pt", "BR"));
        clockText.setText(timeFormat.format(now.getTime()));
        dateText.setText(dateFormat.format(now.getTime()));
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

        fileRepository.getFileTree(1, 10, new ApiCallback<PaginatedResult<FileItem>>() {
            @Override
            public void onSuccess(PaginatedResult<FileItem> result) {
                if (!isAdded()) {
                    return;
                }
                List<FileItem> items = result.getItems();
                if (items != null && !items.isEmpty()) {
                    recentAdapter.updateItems(items);
                    sectionRecent.setVisibility(View.VISIBLE);
                } else {
                    sectionRecent.setVisibility(View.GONE);
                }
                recentLoaded = true;
                checkLoadComplete();
            }

            @Override
            public void onError(AppError error) {
                if (!isAdded()) {
                    return;
                }
                sectionRecent.setVisibility(View.GONE);
                recentLoaded = true;
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
                    sectionStarred.setVisibility(View.VISIBLE);
                } else {
                    sectionStarred.setVisibility(View.GONE);
                }
                starredLoaded = true;
                checkLoadComplete();
            }

            @Override
            public void onError(AppError error) {
                if (!isAdded()) {
                    return;
                }
                sectionStarred.setVisibility(View.GONE);
                starredLoaded = true;
                checkLoadComplete();
            }
        });
    }

    private void checkLoadComplete() {
        if (!recentLoaded || !starredLoaded) {
            return;
        }
        swipeRefresh.setRefreshing(false);
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
