package com.kuranas.mobile.app;

import android.os.Bundle;

import androidx.appcompat.app.AppCompatActivity;

import com.kuranas.mobile.R;
import com.kuranas.mobile.infra.kiosk.KioskManager;
import com.kuranas.mobile.presentation.home.HomeFragment;

public class MainActivity extends AppCompatActivity {

    private KioskManager kioskManager;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        kioskManager = new KioskManager(this);
        kioskManager.engage();

        if (savedInstanceState == null) {
            getSupportFragmentManager()
                    .beginTransaction()
                    .replace(R.id.content_frame, new HomeFragment())
                    .commit();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        kioskManager.engage();
    }
}
