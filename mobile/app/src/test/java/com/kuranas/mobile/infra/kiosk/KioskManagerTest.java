package com.kuranas.mobile.infra.kiosk;

import android.app.Activity;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;

import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import static org.junit.Assert.assertNotNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

public class KioskManagerTest {

    private static final int FULLSCREEN_FLAGS =
            View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_LOW_PROFILE;

    private Activity activity;
    private Window window;
    private View decorView;
    private KioskManager kioskManager;

    @Before
    public void setUp() {
        activity = mock(Activity.class);
        window = mock(Window.class);
        decorView = mock(View.class);

        when(activity.getWindow()).thenReturn(window);
        when(window.getDecorView()).thenReturn(decorView);

        kioskManager = new KioskManager(activity);
    }

    @Test
    public void engage_addsKeepScreenOnFlag() {
        kioskManager.engage();

        verify(window).addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
    }

    @Test
    public void engage_setsFullscreenSystemUiFlags() {
        kioskManager.engage();

        verify(decorView).setSystemUiVisibility(FULLSCREEN_FLAGS);
    }

    @Test
    public void engage_registersSystemUiVisibilityListener() {
        kioskManager.engage();

        verify(decorView).setOnSystemUiVisibilityChangeListener(
                org.mockito.ArgumentMatchers.any(View.OnSystemUiVisibilityChangeListener.class));
    }

    @Test
    public void systemUiListener_reappliesFullscreenWhenBarsVisible() {
        kioskManager.engage();

        ArgumentCaptor<View.OnSystemUiVisibilityChangeListener> captor =
                ArgumentCaptor.forClass(View.OnSystemUiVisibilityChangeListener.class);
        verify(decorView).setOnSystemUiVisibilityChangeListener(captor.capture());

        View.OnSystemUiVisibilityChangeListener listener = captor.getValue();
        assertNotNull(listener);

        // Simulate system bars becoming visible (HIDE_NAVIGATION bit cleared)
        listener.onSystemUiVisibilityChange(0);

        // Should re-apply fullscreen flags (once in engage + once from listener)
        verify(decorView, times(2)).setSystemUiVisibility(FULLSCREEN_FLAGS);
    }

    @Test
    public void systemUiListener_doesNotReapplyWhenBarsHidden() {
        kioskManager.engage();

        ArgumentCaptor<View.OnSystemUiVisibilityChangeListener> captor =
                ArgumentCaptor.forClass(View.OnSystemUiVisibilityChangeListener.class);
        verify(decorView).setOnSystemUiVisibilityChangeListener(captor.capture());

        View.OnSystemUiVisibilityChangeListener listener = captor.getValue();

        // Simulate bars still hidden (HIDE_NAVIGATION bit set)
        listener.onSystemUiVisibilityChange(View.SYSTEM_UI_FLAG_HIDE_NAVIGATION);

        // setSystemUiVisibility should only have been called once (in engage), not again
        verify(decorView).setSystemUiVisibility(FULLSCREEN_FLAGS);
    }
}
