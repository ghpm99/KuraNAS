package com.kuranas.mobile.presentation.images;

import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.fragment.app.Fragment;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.presentation.common.BitmapLoaderTask;

import java.util.ArrayList;

public class ImageViewerFragment extends Fragment {

    private static final String ARG_IMAGE_IDS = "image_ids";
    private static final String ARG_POSITION = "position";
    private static final int IMAGE_SIZE = 1024;

    private ImageView imageView;
    private Button btnPrev;
    private Button btnNext;
    private TextView imageCounter;

    private ArrayList<Integer> imageIds;
    private int currentPosition;
    private String baseUrl;
    private BitmapCache bitmapCache;

    public static ImageViewerFragment newInstance(ArrayList<Integer> imageIds, int position) {
        ImageViewerFragment fragment = new ImageViewerFragment();
        Bundle args = new Bundle();
        args.putIntegerArrayList(ARG_IMAGE_IDS, imageIds);
        args.putInt(ARG_POSITION, position);
        fragment.setArguments(args);
        return fragment;
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_image_viewer, container, false);

        ServiceLocator locator = ServiceLocator.getInstance();
        baseUrl = locator.getHttpClient().getBaseUrl();
        bitmapCache = locator.getBitmapCache();

        imageView = (ImageView) root.findViewById(R.id.image_view);
        btnPrev = (Button) root.findViewById(R.id.btn_prev);
        btnNext = (Button) root.findViewById(R.id.btn_next);
        imageCounter = (TextView) root.findViewById(R.id.image_counter);

        Bundle args = getArguments();
        if (args != null) {
            imageIds = args.getIntegerArrayList(ARG_IMAGE_IDS);
            currentPosition = args.getInt(ARG_POSITION, 0);
        }

        if (imageIds == null) {
            imageIds = new ArrayList<Integer>();
        }

        btnPrev.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (currentPosition > 0) {
                    currentPosition--;
                    displayCurrentImage();
                }
            }
        });

        btnNext.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (currentPosition < imageIds.size() - 1) {
                    currentPosition++;
                    displayCurrentImage();
                }
            }
        });

        displayCurrentImage();

        return root;
    }

    private void displayCurrentImage() {
        if (imageIds.isEmpty()) {
            return;
        }

        if (currentPosition < 0) {
            currentPosition = 0;
        }
        if (currentPosition >= imageIds.size()) {
            currentPosition = imageIds.size() - 1;
        }

        int imageId = imageIds.get(currentPosition).intValue();
        String imageUrl = baseUrl + "/api/v1/files/blob/" + imageId;

        BitmapLoaderTask.load(imageUrl, imageView, bitmapCache, IMAGE_SIZE, IMAGE_SIZE);

        String counterText = (currentPosition + 1) + " / " + imageIds.size();
        imageCounter.setText(counterText);

        btnPrev.setEnabled(currentPosition > 0);
        btnNext.setEnabled(currentPosition < imageIds.size() - 1);
    }
}
