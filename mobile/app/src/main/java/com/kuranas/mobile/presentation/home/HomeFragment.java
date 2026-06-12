package com.kuranas.mobile.presentation.home;

import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.presentation.base.BaseFragment;

import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.Locale;

/**
 * The single screen of the wall-panel app: a big clock and the date. Task 18
 * builds the kiosk sections (notifications, e-mail digest) on top of it.
 */
public class HomeFragment extends BaseFragment {

    private static final long CLOCK_UPDATE_INTERVAL_MS = 15000;

    private TextView clockText;
    private TextView dateText;

    private final Handler clockHandler = new Handler(Looper.getMainLooper());
    private final Runnable clockRunnable = new Runnable() {
        @Override
        public void run() {
            updateClock();
            clockHandler.postDelayed(this, CLOCK_UPDATE_INTERVAL_MS);
        }
    };

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_home, container, false);

        clockText = (TextView) root.findViewById(R.id.clock_text);
        dateText = (TextView) root.findViewById(R.id.date_text);

        updateClock();

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
}
