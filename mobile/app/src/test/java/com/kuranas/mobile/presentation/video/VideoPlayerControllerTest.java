package com.kuranas.mobile.presentation.video;

import com.kuranas.mobile.domain.model.VideoPlaybackState;

import org.junit.Before;
import org.junit.Test;

import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

public class VideoPlayerControllerTest {

    private FakeRepository repository;
    private FakeEngine engine;
    private FakeView view;
    private FakeScheduler scheduler;
    private VideoPlayerController controller;

    @Before
    public void setUp() {
        repository = new FakeRepository();
        engine = new FakeEngine();
        view = new FakeView();
        scheduler = new FakeScheduler();
        controller = new VideoPlayerController(repository, engine, view, scheduler);
    }

    @Test
    public void start_withStoredState_preparesAndStartsPlayback() {
        engine.duration = 120000;
        controller.start(7, "http://stream");

        assertEquals(7, repository.startedVideoId);

        VideoPlaybackState state = new VideoPlaybackState(
                1, "client", 0, 7, 12.5, 120.0, false, false, "2026-04-01T00:00:00Z"
        );
        repository.succeedStart(state);
        engine.triggerPrepared();

        assertEquals("http://stream", engine.videoUri);
        assertEquals(120000, view.seekMax);
        assertEquals(12500, engine.lastSeekPosition);
        assertTrue(view.playPausePlaying);
        assertEquals("0:12 / 2:00", view.timeText);
        assertTrue(scheduler.hasDelayed(VideoPlayerController.CONTROLS_HIDE_DELAY_MS));
        assertTrue(scheduler.hasDelayed(VideoPlayerController.STATE_UPDATE_INTERVAL_MS));
        assertEquals(1, scheduler.postCalls);
    }

    @Test
    public void start_whenRepositoryFails_startsFromBeginning() {
        engine.duration = 60000;

        controller.start(8, "http://stream-2");
        repository.failStart();
        engine.triggerPrepared();

        assertEquals("http://stream-2", engine.videoUri);
        assertEquals(-1, engine.lastSeekPosition);
        assertTrue(view.playPausePlaying);
    }

    @Test
    public void onPlayPauseClicked_whenPlaying_pausesAndSavesState() {
        engine.duration = 60000;
        engine.current = 5000;
        prepareStartedController();

        controller.onPlayPauseClicked();

        assertEquals(1, engine.pauseCalls);
        assertFalse(view.playPausePlaying);
        assertEquals(1, repository.updateCalls);
        assertTrue(repository.lastPaused);
        assertFalse(repository.lastCompleted);
    }

    @Test
    public void onPlayPauseClicked_whenPaused_resumesAndSchedulesHide() {
        engine.duration = 60000;
        prepareStartedController();
        controller.onPlayPauseClicked();

        int startCallsBeforeResume = engine.startCalls;
        int postCallsBeforeResume = scheduler.postCalls;

        controller.onPlayPauseClicked();

        assertEquals(startCallsBeforeResume + 1, engine.startCalls);
        assertTrue(view.playPausePlaying);
        assertEquals(postCallsBeforeResume + 1, scheduler.postCalls);
        assertTrue(scheduler.hasDelayed(VideoPlayerController.CONTROLS_HIDE_DELAY_MS));
    }

    @Test
    public void onVideoTapped_togglesOverlayVisibility() {
        view.controlsVisible = true;

        controller.onVideoTapped();
        assertFalse(view.controlsVisible);

        controller.onVideoTapped();
        assertTrue(view.controlsVisible);
        assertTrue(scheduler.hasDelayed(VideoPlayerController.CONTROLS_HIDE_DELAY_MS));
    }

    @Test
    public void onSeekCallbacks_handleUserSeekAndHideScheduling() {
        engine.duration = 10000;
        controller.onSeekProgressChanged(3000, false);
        assertEquals(-1, engine.lastSeekPosition);

        controller.onSeekProgressChanged(3000, true);
        assertEquals(3000, engine.lastSeekPosition);
        assertEquals("0:03 / 0:10", view.timeText);

        int removeCallsBefore = scheduler.removeCalls;
        controller.onSeekStartTracking();
        assertEquals(removeCallsBefore + 1, scheduler.removeCalls);

        controller.onSeekStopTracking();
        assertTrue(scheduler.hasDelayed(VideoPlayerController.CONTROLS_HIDE_DELAY_MS));
    }

    @Test
    public void onPause_whenPlaying_savesStateAndStopsSchedulers() {
        engine.duration = 60000;
        engine.current = 2000;
        prepareStartedController();

        controller.onPause();

        assertEquals(1, engine.pauseCalls);
        assertEquals(1, repository.updateCalls);
        assertTrue(repository.lastPaused);
        assertFalse(repository.lastCompleted);
        assertTrue(scheduler.removeCalls >= 3);
    }

    @Test
    public void onResume_whenControllerStillMarkedPlaying_restartsPlayback() {
        engine.duration = 30000;
        prepareStartedController();
        engine.playing = false;

        int startCallsBeforePause = engine.startCalls;
        controller.onPause();

        controller.onResume();

        assertEquals(startCallsBeforePause + 1, engine.startCalls);
        assertTrue(scheduler.postCalls >= 2);
        assertTrue(scheduler.hasDelayed(VideoPlayerController.STATE_UPDATE_INTERVAL_MS));
    }

    @Test
    public void completionCallback_showsControlsAndSavesCompletedState() {
        engine.duration = 40000;
        prepareStartedController();

        engine.triggerCompletion();

        assertFalse(view.playPausePlaying);
        assertTrue(view.controlsVisible);
        assertEquals(1, repository.updateCalls);
        assertTrue(repository.lastCompleted);
    }

    @Test
    public void onDestroy_stopsEngineAndClearsScheduler() {
        controller.onDestroy();

        assertEquals(1, engine.stopCalls);
        assertTrue(scheduler.cleared);
    }

    @Test
    public void formatTime_handlesHoursAndNegativeValues() {
        assertEquals("1:01:01", controller.formatTime(3661000));
        assertEquals("0:00", controller.formatTime(-1));
    }

    private void prepareStartedController() {
        controller.start(9, "http://stream");
        repository.succeedStart(null);
        engine.triggerPrepared();
    }

    private static final class FakeRepository implements VideoPlayerController.PlaybackStateRepository {
        private int startedVideoId;
        private StartPlaybackCallback callback;
        private int updateCalls;
        private int lastUpdatedVideoId;
        private double lastCurrentTimeSec;
        private double lastDurationSec;
        private boolean lastPaused;
        private boolean lastCompleted;

        @Override
        public void startPlayback(int videoId, StartPlaybackCallback callback) {
            this.startedVideoId = videoId;
            this.callback = callback;
        }

        @Override
        public void updatePlaybackState(
                int videoId,
                double currentTimeSec,
                double durationSec,
                boolean paused,
                boolean completed
        ) {
            updateCalls++;
            lastUpdatedVideoId = videoId;
            lastCurrentTimeSec = currentTimeSec;
            lastDurationSec = durationSec;
            lastPaused = paused;
            lastCompleted = completed;
        }

        private void succeedStart(VideoPlaybackState state) {
            callback.onSuccess(state);
        }

        private void failStart() {
            callback.onError();
        }
    }

    private static final class FakeEngine implements VideoPlayerController.PlaybackEngine {
        private String videoUri;
        private Runnable onPrepared;
        private Runnable onCompletion;
        private int duration;
        private int current;
        private boolean playing;
        private int lastSeekPosition = -1;
        private int startCalls;
        private int pauseCalls;
        private int stopCalls;

        @Override
        public void setVideoUri(String streamUrl) {
            this.videoUri = streamUrl;
        }

        @Override
        public void setOnPrepared(Runnable runnable) {
            this.onPrepared = runnable;
        }

        @Override
        public void setOnCompletion(Runnable runnable) {
            this.onCompletion = runnable;
        }

        @Override
        public int getDuration() {
            return duration;
        }

        @Override
        public int getCurrentPosition() {
            return current;
        }

        @Override
        public boolean isPlaying() {
            return playing;
        }

        @Override
        public void seekTo(int positionMs) {
            lastSeekPosition = positionMs;
            current = positionMs;
        }

        @Override
        public void start() {
            playing = true;
            startCalls++;
        }

        @Override
        public void pause() {
            playing = false;
            pauseCalls++;
        }

        @Override
        public void stop() {
            stopCalls++;
        }

        private void triggerPrepared() {
            if (onPrepared != null) {
                onPrepared.run();
            }
        }

        private void triggerCompletion() {
            playing = false;
            if (onCompletion != null) {
                onCompletion.run();
            }
        }
    }

    private static final class FakeView implements VideoPlayerController.ViewContract {
        private boolean playPausePlaying;
        private boolean controlsVisible = true;
        private int seekMax;
        private int seekProgress;
        private String timeText;

        @Override
        public void setPlayPausePlaying(boolean playing) {
            this.playPausePlaying = playing;
        }

        @Override
        public void setControlsVisible(boolean visible) {
            this.controlsVisible = visible;
        }

        @Override
        public boolean isControlsVisible() {
            return controlsVisible;
        }

        @Override
        public void setSeekMax(int max) {
            this.seekMax = max;
        }

        @Override
        public void setSeekProgress(int progress) {
            this.seekProgress = progress;
        }

        @Override
        public void setTimeText(String text) {
            this.timeText = text;
        }
    }

    private static final class FakeScheduler implements VideoPlayerController.Scheduler {
        private final List<Runnable> postedTasks = new ArrayList<Runnable>();
        private final List<DelayedTask> delayedTasks = new ArrayList<DelayedTask>();
        private int postCalls;
        private int removeCalls;
        private boolean cleared;

        @Override
        public void post(Runnable runnable) {
            postCalls++;
            postedTasks.add(runnable);
        }

        @Override
        public void postDelayed(Runnable runnable, long delayMs) {
            delayedTasks.add(new DelayedTask(runnable, delayMs));
        }

        @Override
        public void removeCallbacks(Runnable runnable) {
            removeCalls++;
            postedTasks.remove(runnable);
            removeFromDelayed(runnable);
        }

        @Override
        public void clear() {
            cleared = true;
            postedTasks.clear();
            delayedTasks.clear();
        }

        private boolean hasDelayed(long delayMs) {
            for (DelayedTask task : delayedTasks) {
                if (task.delayMs == delayMs) {
                    return true;
                }
            }
            return false;
        }

        private void removeFromDelayed(Runnable runnable) {
            Iterator<DelayedTask> iterator = delayedTasks.iterator();
            while (iterator.hasNext()) {
                DelayedTask task = iterator.next();
                if (task.runnable == runnable) {
                    iterator.remove();
                }
            }
        }

        private static final class DelayedTask {
            private final Runnable runnable;
            private final long delayMs;

            private DelayedTask(Runnable runnable, long delayMs) {
                this.runnable = runnable;
                this.delayMs = delayMs;
            }
        }
    }
}
