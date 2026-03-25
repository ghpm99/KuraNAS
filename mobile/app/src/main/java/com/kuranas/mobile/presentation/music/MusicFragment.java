package com.kuranas.mobile.presentation.music;

import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageButton;
import android.widget.TextView;

import androidx.fragment.app.FragmentActivity;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import androidx.swiperefreshlayout.widget.SwipeRefreshLayout;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Track;
import com.kuranas.mobile.domain.repository.MusicRepository;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;
import com.kuranas.mobile.presentation.base.ViewState;

import java.util.ArrayList;
import java.util.List;

public class MusicFragment extends BaseFragment {

    private static final int PAGE_SIZE = 30;

    private SwipeRefreshLayout swipeRefresh;
    private RecyclerView tracksList;
    private View playerBar;
    private TextView playerTitle;
    private TextView playerArtist;
    private ImageButton btnPlayPause;

    private MusicAdapter adapter;
    private MusicRepository musicRepository;
    private TranslationManager translations;

    private final List<Track> allTracks = new ArrayList<Track>();
    private int currentPage = 1;
    private boolean hasNextPage = false;
    private boolean isLoadingMore = false;
    private int currentTrackFileId = -1;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_music, container, false);

        initStateViews(root);

        ServiceLocator locator = ServiceLocator.getInstance();
        musicRepository = locator.getMusicRepository();
        translations = locator.getTranslationManager();

        swipeRefresh = (SwipeRefreshLayout) root.findViewById(R.id.swipe_refresh);
        tracksList = (RecyclerView) root.findViewById(R.id.tracks_list);
        playerBar = root.findViewById(R.id.player_bar);
        playerTitle = (TextView) root.findViewById(R.id.player_title);
        playerArtist = (TextView) root.findViewById(R.id.player_artist);
        btnPlayPause = (ImageButton) root.findViewById(R.id.btn_play_pause);

        setupRecyclerView();

        if (playerBar != null) {
            playerBar.setVisibility(View.GONE);
        }

        setRetryListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                loadTracks(true);
            }
        });

        swipeRefresh.setOnRefreshListener(new SwipeRefreshLayout.OnRefreshListener() {
            @Override
            public void onRefresh() {
                loadTracks(true);
            }
        });

        loadTracks(true);

        return root;
    }

    @Override
    public void onResume() {
        super.onResume();
        loadPlayerState();
    }

    private void setupRecyclerView() {
        final LinearLayoutManager layoutManager = new LinearLayoutManager(getActivity());
        tracksList.setLayoutManager(layoutManager);

        adapter = new MusicAdapter(new ArrayList<Track>(), new MusicAdapter.OnTrackClickListener() {
            @Override
            public void onTrackClick(Track track, int position) {
                handleTrackClick(track, position);
            }
        }, translations);

        tracksList.setAdapter(adapter);

        tracksList.addOnScrollListener(new RecyclerView.OnScrollListener() {
            @Override
            public void onScrolled(RecyclerView recyclerView, int dx, int dy) {
                super.onScrolled(recyclerView, dx, dy);
                if (isLoadingMore || !hasNextPage) {
                    return;
                }
                int totalItemCount = layoutManager.getItemCount();
                int lastVisibleItem = layoutManager.findLastVisibleItemPosition();
                if (lastVisibleItem >= totalItemCount - 5) {
                    loadMoreTracks();
                }
            }
        });
    }

    private void loadTracks(boolean reset) {
        if (reset) {
            currentPage = 1;
            allTracks.clear();
        }
        if (currentPage == 1) {
            setState(ViewState.LOADING);
        }
        isLoadingMore = currentPage > 1;

        musicRepository.getLibraryTracks(currentPage, PAGE_SIZE,
                new ApiCallback<PaginatedResult<Track>>() {
                    @Override
                    public void onSuccess(PaginatedResult<Track> result) {
                        if (!isAdded()) {
                            return;
                        }
                        swipeRefresh.setRefreshing(false);
                        isLoadingMore = false;
                        hasNextPage = result.hasNext();

                        List<Track> items = result.getItems();
                        if (currentPage == 1) {
                            if (items != null && !items.isEmpty()) {
                                allTracks.addAll(items);
                                adapter.updateItems(items);
                                setState(ViewState.CONTENT);
                            } else {
                                setEmptyMessage(t("MUSIC_PLAYLIST_EMPTY"));
                                setState(ViewState.EMPTY);
                            }
                        } else {
                            if (items != null && !items.isEmpty()) {
                                allTracks.addAll(items);
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

    private void loadMoreTracks() {
        currentPage++;
        loadTracks(false);
    }

    private void loadPlayerState() {
        musicRepository.getPlayerState(new ApiCallback<MusicPlayerState>() {
            @Override
            public void onSuccess(MusicPlayerState state) {
                if (!isAdded()) {
                    return;
                }
                if (state != null && state.getCurrentFileId() > 0) {
                    currentTrackFileId = state.getCurrentFileId();
                    adapter.setCurrentTrackId(currentTrackFileId);
                    updateMiniPlayer(state);
                }
            }

            @Override
            public void onError(AppError error) {
                // Mini player remains hidden
            }
        });
    }

    private void handleTrackClick(Track track, int position) {
        currentTrackFileId = track.getFileId();
        adapter.setCurrentTrackId(currentTrackFileId);
        updateMiniPlayerFromTrack(track);
        openMusicPlayer(track);
    }

    private void openMusicPlayer(Track track) {
        MusicPlayerFragment playerFragment = MusicPlayerFragment.newInstance(
                track.getFileId(), track.getDisplayTitle(), track.getDisplayArtist());

        FragmentActivity activity = getActivity();
        if (activity != null) {
            activity.getSupportFragmentManager()
                    .beginTransaction()
                    .replace(R.id.content_frame, playerFragment)
                    .addToBackStack(null)
                    .commit();
        }
    }

    private void updateMiniPlayer(MusicPlayerState state) {
        if (playerBar == null) {
            return;
        }
        playerBar.setVisibility(View.VISIBLE);

        // Find the track matching the state
        for (int i = 0; i < allTracks.size(); i++) {
            Track track = allTracks.get(i);
            if (track.getFileId() == state.getCurrentFileId()) {
                playerTitle.setText(track.getDisplayTitle());
                playerArtist.setText(track.getDisplayArtist());
                break;
            }
        }

        if (btnPlayPause != null) {
            btnPlayPause.setOnClickListener(new View.OnClickListener() {
                @Override
                public void onClick(View v) {
                    // Find current track and open player
                    for (int i = 0; i < allTracks.size(); i++) {
                        Track track = allTracks.get(i);
                        if (track.getFileId() == currentTrackFileId) {
                            openMusicPlayer(track);
                            break;
                        }
                    }
                }
            });
        }
    }

    private void updateMiniPlayerFromTrack(Track track) {
        if (playerBar == null) {
            return;
        }
        playerBar.setVisibility(View.VISIBLE);
        playerTitle.setText(track.getDisplayTitle());
        playerArtist.setText(track.getDisplayArtist());
    }
}
