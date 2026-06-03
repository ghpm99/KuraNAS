package com.kuranas.mobile.app;

import android.content.Intent;
import android.os.Bundle;
import android.view.View;
import android.view.ViewGroup;
import android.widget.AdapterView;
import android.widget.BaseAdapter;
import android.widget.ImageView;
import android.widget.ListView;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;
import androidx.drawerlayout.widget.DrawerLayout;
import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentTransaction;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.feature.files.presentation.FilesFragment;
import com.kuranas.mobile.feature.images.presentation.ImagesFragment;
import com.kuranas.mobile.feature.search.presentation.SearchFragment;
import com.kuranas.mobile.feature.settings.presentation.SettingsFragment;
import com.kuranas.mobile.infra.kiosk.KioskManager;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.presentation.home.HomeFragment;
import com.kuranas.mobile.presentation.images.ImageViewerFragment;
import com.kuranas.mobile.presentation.music.MusicFragment;
import com.kuranas.mobile.presentation.music.MusicPlayerFragment;
import com.kuranas.mobile.presentation.video.VideoFragment;
import com.kuranas.mobile.presentation.video.VideoPlayerActivity;

import java.util.ArrayList;

public class MainActivity extends AppCompatActivity
        implements HomeFragment.NavigationHost,
        com.kuranas.mobile.presentation.files.FilesFragment.FileNavigationHost,
        com.kuranas.mobile.presentation.search.SearchFragment.SearchNavigationHost {

    private final MainNavigationCoordinator navigationCoordinator = new MainNavigationCoordinator();
    private KioskManager kioskManager;
    private DrawerLayout drawerLayout;
    private View navDrawer;
    private ListView navList;
    private int currentNavPosition = 0;

    private final int[] navIcons = {
            R.drawable.ic_home,
            R.drawable.ic_folder,
            R.drawable.ic_image,
            R.drawable.ic_music,
            R.drawable.ic_video,
            R.drawable.ic_search,
            R.drawable.ic_settings
    };

    private String[] navLabels;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        kioskManager = new KioskManager(this);
        kioskManager.engage();

        drawerLayout = (DrawerLayout) findViewById(R.id.drawer_layout);
        navDrawer = findViewById(R.id.nav_drawer);
        navList = (ListView) findViewById(R.id.nav_list);
        navLabels = buildNavLabels();

        setupNavDrawer();

        if (savedInstanceState == null) {
            navigateTo(MainNavigationCoordinator.NAV_HOME);
        }
    }

    private String[] buildNavLabels() {
        return new String[]{
                getString(R.string.nav_home),
                getString(R.string.nav_files),
                getString(R.string.nav_images),
                getString(R.string.nav_music),
                getString(R.string.nav_videos),
                getString(R.string.nav_search),
                getString(R.string.nav_settings)
        };
    }

    private void setupNavDrawer() {
        navList.setAdapter(new NavAdapter());
        navList.setOnItemClickListener(new AdapterView.OnItemClickListener() {
            @Override
            public void onItemClick(AdapterView<?> parent, View view, int position, long id) {
                navigateTo(position);
                drawerLayout.closeDrawers();
            }
        });
    }

    private void navigateTo(int position) {
        MainNavigationCoordinator.NavigationInstruction instruction =
                navigationCoordinator.resolveNavigation(position);
        currentNavPosition = instruction.getSelectedPosition();
        Fragment fragment = createFragmentForDestination(instruction.getDestination());

        FragmentTransaction tx = getSupportFragmentManager().beginTransaction();
        tx.replace(R.id.content_frame, fragment);
        if (instruction.shouldAddToBackStack()) {
            tx.addToBackStack(null);
        }
        tx.commit();

        navList.setItemChecked(instruction.getSelectedPosition(), true);
    }

    private Fragment createFragmentForDestination(MainNavigationCoordinator.Destination destination) {
        switch (destination) {
            case FILES:
                return new FilesFragment();
            case IMAGES:
                return new ImagesFragment();
            case MUSIC:
                return new MusicFragment();
            case VIDEOS:
                return new VideoFragment();
            case SEARCH:
                return new SearchFragment();
            case SETTINGS:
                return new SettingsFragment();
            case HOME:
            default:
                return new HomeFragment();
        }
    }

    @Override
    public void onBackPressed() {
        Fragment current = getSupportFragmentManager().findFragmentById(R.id.content_frame);
        boolean filesHandledBackNavigation = false;
        if (current instanceof FilesFragment) {
            filesHandledBackNavigation = ((FilesFragment) current).handleBackNavigation();
        }

        boolean isDrawerOpen =
                drawerLayout != null && navDrawer != null && drawerLayout.isDrawerOpen(navDrawer);
        MainNavigationCoordinator.BackPressDecision decision = navigationCoordinator.resolveBackPress(
                isDrawerOpen,
                filesHandledBackNavigation,
                getSupportFragmentManager().getBackStackEntryCount()
        );

        switch (decision.getAction()) {
            case CLOSE_DRAWER:
                if (drawerLayout != null && navDrawer != null) {
                    drawerLayout.closeDrawer(navDrawer);
                }
                return;
            case CONSUMED:
                return;
            case POP_BACK_STACK:
                getSupportFragmentManager().popBackStack();
                if (decision.shouldUpdateSelection()) {
                    currentNavPosition = decision.getSelectedPosition();
                    navList.setItemChecked(currentNavPosition, true);
                }
                return;
            case NO_OP:
            default:
                return;
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        kioskManager.engage();
    }

    // HomeFragment.NavigationHost
    @Override
    public void onFileItemSelected(FileItem item) {
        if (item.isDirectory()) {
            FilesFragment fragment = new FilesFragment();
            Bundle args = new Bundle();
            args.putString("path", item.getPath());
            fragment.setArguments(args);
            showFragment(fragment);
        } else if (item.isImage()) {
            openImageViewer(item);
        } else if (item.isAudio()) {
            openMusicPlayer(item);
        } else if (item.isVideo()) {
            openVideoPlayer(item);
        }
    }

    // FilesFragment.FileNavigationHost
    @Override
    public void openImageViewer(FileItem item) {
        ArrayList<Integer> ids = new ArrayList<Integer>();
        ids.add(item.getId());
        ImageViewerFragment fragment = ImageViewerFragment.newInstance(ids, 0);
        showFragment(fragment);
    }

    @Override
    public void openMusicPlayer(FileItem item) {
        MusicPlayerFragment fragment = new MusicPlayerFragment();
        Bundle args = new Bundle();
        args.putInt("fileId", item.getId());
        args.putString("title", item.getName());
        args.putString("artist", "");
        fragment.setArguments(args);
        showFragment(fragment);
    }

    @Override
    public void openVideoPlayer(FileItem item) {
        String baseUrl = ServiceLocator.getInstance().getHttpClient().getBaseUrl();
        Intent intent = new Intent(this, VideoPlayerActivity.class);
        intent.putExtra("videoId", item.getId());
        intent.putExtra("videoName", item.getName());
        intent.putExtra("streamUrl", baseUrl + "/api/v1/files/video-stream/" + item.getId());
        startActivity(intent);
    }

    @Override
    public void openFile(FileItem item) {
        // Generic file - navigate to parent directory
        FilesFragment fragment = new FilesFragment();
        Bundle args = new Bundle();
        args.putString("path", item.getParentPath());
        fragment.setArguments(args);
        showFragment(fragment);
    }

    // SearchFragment.SearchNavigationHost
    @Override
    public void onSearchFileSelected(FileItem file) {
        onFileItemSelected(file);
    }

    @Override
    public void onSearchFolderSelected(FileItem folder) {
        FilesFragment fragment = new FilesFragment();
        Bundle args = new Bundle();
        args.putString("path", folder.getPath());
        fragment.setArguments(args);
        showFragment(fragment);
    }

    @Override
    public void onSearchImageSelected(FileItem image) {
        openImageViewer(image);
    }

    @Override
    public void onSearchVideoSelected(VideoItem video) {
        String baseUrl = ServiceLocator.getInstance().getHttpClient().getBaseUrl();
        Intent intent = new Intent(this, VideoPlayerActivity.class);
        intent.putExtra("videoId", video.getId());
        intent.putExtra("videoName", video.getName());
        intent.putExtra("streamUrl", baseUrl + "/api/v1/files/video-stream/" + video.getId());
        startActivity(intent);
    }

    private void showFragment(Fragment fragment) {
        FragmentTransaction tx = getSupportFragmentManager().beginTransaction();
        tx.replace(R.id.content_frame, fragment);
        tx.addToBackStack(null);
        tx.commit();
    }

    private class NavAdapter extends BaseAdapter {

        @Override
        public int getCount() {
            return navLabels.length;
        }

        @Override
        public Object getItem(int position) {
            return navLabels[position];
        }

        @Override
        public long getItemId(int position) {
            return position;
        }

        @Override
        public View getView(int position, View convertView, ViewGroup parent) {
            if (convertView == null) {
                convertView = getLayoutInflater().inflate(R.layout.item_nav_drawer, parent, false);
            }

            ImageView icon = (ImageView) convertView.findViewById(R.id.nav_icon);
            TextView label = (TextView) convertView.findViewById(R.id.nav_label);

            icon.setImageResource(navIcons[position]);
            label.setText(navLabels[position]);

            if (position == currentNavPosition) {
                convertView.setBackgroundColor(getResources().getColor(R.color.nav_drawer_selected));
            } else {
                convertView.setBackgroundColor(getResources().getColor(R.color.transparent));
            }

            return convertView;
        }
    }
}
