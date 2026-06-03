package com.kuranas.mobile.presentation.music;

import org.junit.Before;
import org.junit.Test;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

public class MusicPlayerControllerTest {

    private FakeView view;
    private FakeAudioEngine audioEngine;
    private FakeRepository repository;
    private FakeScheduler scheduler;
    private MusicPlayerController controller;

    @Before
    public void setUp() {
        view = new FakeView();
        audioEngine = new FakeAudioEngine();
        repository = new FakeRepository();
        scheduler = new FakeScheduler();
        controller = new MusicPlayerController(view, audioEngine, repository, scheduler);
    }

    @Test
    public void start_withValidState_preparesAndStartsPlayback() {
        audioEngine.duration = 90000;
        MusicPlayerController.PlayerState state =
                new MusicPlayerController.PlayerState(11, "Track", "Artist");

        controller.start("http://base", state);
        audioEngine.triggerPrepared();

        assertEquals("Track", view.title);
        assertEquals("Artist", view.artist);
        assertTrue(view.defaultAlbumArtSet);
        assertTrue(audioEngine.streamTypeSet);
        assertEquals("http://base/api/v1/files/stream/11", audioEngine.dataSource);
        assertTrue(view.playPausePlaying);
        assertEquals(90000, view.seekMax);
        assertEquals("1:30", view.totalTimeText);
        assertEquals(1, scheduler.postCalls);
    }

    @Test
    public void start_whenDataSourceFails_doesNotPrepareAsync() {
        audioEngine.throwOnDataSource = true;

        controller.start("http://base", new MusicPlayerController.PlayerState(1, "A", "B"));

        assertEquals(0, audioEngine.prepareAsyncCalls);
    }

    @Test
    public void onPlayPauseClicked_whenNotPrepared_doesNothing() {
        controller.onPlayPauseClicked();

        assertEquals(0, audioEngine.pauseCalls);
        assertEquals(0, audioEngine.startCalls);
    }

    @Test
    public void onSeekProgressChanged_whenNotFromUser_doesNotUpdateCurrentTime() {
        prepareStartedController();
        view.currentTimeText = null;

        controller.onSeekProgressChanged(5000, false);

        assertEquals(null, view.currentTimeText);
    }

    @Test
    public void onPlayPauseClicked_togglesPauseAndResume() {
        prepareStartedController();
        audioEngine.currentPosition = 4000;

        controller.onPlayPauseClicked();

        assertEquals(1, audioEngine.pauseCalls);
        assertFalse(view.playPausePlaying);
        assertEquals(1, repository.updateCalls);
        assertEquals(4.0, repository.lastPositionSeconds, 0.001);

        int startCallsBeforeResume = audioEngine.startCalls;
        controller.onPlayPauseClicked();

        assertEquals(startCallsBeforeResume + 1, audioEngine.startCalls);
        assertTrue(view.playPausePlaying);
    }

    @Test
    public void onPreviousClicked_seeksToStart() {
        prepareStartedController();
        audioEngine.currentPosition = 10000;

        controller.onPreviousClicked();

        assertEquals(0, audioEngine.lastSeekPosition);
        assertEquals(0, view.seekProgress);
        assertEquals("0:00", view.currentTimeText);
    }

    @Test
    public void seekCallbacks_updateSeekingStateAndPlayerPosition() {
        prepareStartedController();

        controller.onSeekProgressChanged(7000, true);
        assertEquals("0:07", view.currentTimeText);

        controller.onSeekStartTracking();
        scheduler.runFirstPostedTask();
        assertEquals(0, view.seekProgress);

        controller.onSeekStopTracking(7000);
        assertEquals(7000, audioEngine.lastSeekPosition);
    }

    @Test
    public void onPause_whenPlaying_pausesAndSavesState() {
        prepareStartedController();
        audioEngine.currentPosition = 3000;

        controller.onPause();

        assertEquals(1, audioEngine.pauseCalls);
        assertFalse(view.playPausePlaying);
        assertEquals(1, repository.updateCalls);
        assertEquals(3.0, repository.lastPositionSeconds, 0.001);
    }

    @Test
    public void onResume_whenPrepared_startsAndSchedulesUpdates() {
        prepareStartedController();
        audioEngine.playing = false;

        int startCallsBeforeResume = audioEngine.startCalls;
        int postCallsBeforeResume = scheduler.postCalls;

        controller.onResume();

        assertEquals(startCallsBeforeResume + 1, audioEngine.startCalls);
        assertTrue(view.playPausePlaying);
        assertEquals(postCallsBeforeResume + 1, scheduler.postCalls);
    }

    @Test
    public void completionCallback_savesStateAndStopsUpdates() {
        prepareStartedController();
        audioEngine.currentPosition = 8000;

        controller.start("http://base", new MusicPlayerController.PlayerState(12, "Next", "Artist"));
        audioEngine.triggerPrepared();
        audioEngine.triggerCompletion();

        assertFalse(view.playPausePlaying);
        assertTrue(repository.updateCalls >= 1);
    }

    @Test
    public void onDestroy_releasesEngineAndClearsScheduler() {
        prepareStartedController();

        controller.onDestroy();

        assertEquals(1, audioEngine.releaseCalls);
        assertTrue(scheduler.cleared);
        assertTrue(repository.updateCalls >= 1);
    }

    @Test
    public void errorCallback_marksPlaybackAsNotPrepared() {
        prepareStartedController();
        int startCallsBefore = audioEngine.startCalls;

        audioEngine.triggerError(1, 2);
        controller.onPlayPauseClicked();

        assertEquals(startCallsBefore, audioEngine.startCalls);
    }

    @Test
    public void formatTime_formatsMinutesAndNormalizesNegative() {
        assertEquals("2:05", controller.formatTime(125000));
        assertEquals("0:00", controller.formatTime(-1));
    }

    private void prepareStartedController() {
        audioEngine.duration = 60000;
        controller.start("http://base", new MusicPlayerController.PlayerState(9, "Song", "Artist"));
        audioEngine.triggerPrepared();
    }

    private static final class FakeView implements MusicPlayerController.ViewContract {
        private String title;
        private String artist;
        private boolean defaultAlbumArtSet;
        private int seekMax;
        private int seekProgress;
        private String currentTimeText;
        private String totalTimeText;
        private boolean playPausePlaying;

        @Override
        public void setTrackTitle(String title) {
            this.title = title;
        }

        @Override
        public void setTrackArtist(String artist) {
            this.artist = artist;
        }

        @Override
        public void setDefaultAlbumArt() {
            defaultAlbumArtSet = true;
        }

        @Override
        public void setSeekMax(int max) {
            seekMax = max;
        }

        @Override
        public void setSeekProgress(int progress) {
            seekProgress = progress;
        }

        @Override
        public void setCurrentTimeText(String text) {
            currentTimeText = text;
        }

        @Override
        public void setTotalTimeText(String text) {
            totalTimeText = text;
        }

        @Override
        public void setPlayPausePlaying(boolean playing) {
            playPausePlaying = playing;
        }
    }

    private static final class FakeAudioEngine implements MusicPlayerController.AudioEngine {
        private boolean streamTypeSet;
        private boolean throwOnDataSource;
        private String dataSource;
        private Runnable onPrepared;
        private Runnable onCompletion;
        private ErrorCallback onError;
        private int prepareAsyncCalls;
        private boolean playing;
        private int duration;
        private int currentPosition;
        private int lastSeekPosition = -1;
        private int startCalls;
        private int pauseCalls;
        private int releaseCalls;

        @Override
        public void setAudioStreamTypeMusic() {
            streamTypeSet = true;
        }

        @Override
        public void setDataSource(String streamUrl) throws IOException {
            if (throwOnDataSource) {
                throw new IOException("expected test failure");
            }
            dataSource = streamUrl;
        }

        @Override
        public void setOnPrepared(Runnable runnable) {
            onPrepared = runnable;
        }

        @Override
        public void setOnCompletion(Runnable runnable) {
            onCompletion = runnable;
        }

        @Override
        public void setOnError(ErrorCallback callback) {
            onError = callback;
        }

        @Override
        public void prepareAsync() {
            prepareAsyncCalls++;
        }

        @Override
        public boolean isPlaying() {
            return playing;
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
        public void seekTo(int positionMs) {
            lastSeekPosition = positionMs;
            currentPosition = positionMs;
        }

        @Override
        public int getDuration() {
            return duration;
        }

        @Override
        public int getCurrentPosition() {
            return currentPosition;
        }

        @Override
        public void release() {
            releaseCalls++;
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

        private void triggerError(int what, int extra) {
            if (onError != null) {
                onError.onError(what, extra);
            }
        }
    }

    private static final class FakeRepository implements MusicPlayerController.PlayerStateRepository {
        private int updateCalls;
        private int lastPlaylistId;
        private int lastFileId;
        private double lastPositionSeconds;

        @Override
        public void updatePlayerState(int playlistId, int fileId, double positionSeconds) {
            updateCalls++;
            lastPlaylistId = playlistId;
            lastFileId = fileId;
            lastPositionSeconds = positionSeconds;
        }
    }

    private static final class FakeScheduler implements MusicPlayerController.Scheduler {
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
            Iterator<DelayedTask> iterator = delayedTasks.iterator();
            while (iterator.hasNext()) {
                DelayedTask task = iterator.next();
                if (task.runnable == runnable) {
                    iterator.remove();
                }
            }
        }

        @Override
        public void clear() {
            cleared = true;
            postedTasks.clear();
            delayedTasks.clear();
        }

        private void runFirstPostedTask() {
            if (postedTasks.isEmpty()) {
                return;
            }
            Runnable task = postedTasks.remove(0);
            task.run();
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
