import {
    clampSeekTime,
    runPlayerShortcut,
    type PlayerShortcutControls,
    type PlayerShortcutSnapshot,
} from './playerShortcutExecution';

const buildControls = (): jest.Mocked<PlayerShortcutControls> => ({
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
    openHelp: jest.fn(),
});

const buildSnapshot = (
    overrides: Partial<PlayerShortcutSnapshot> = {}
): PlayerShortcutSnapshot => ({
    isPlaying: true,
    currentTime: 50,
    duration: 100,
    volume: 0.5,
    playbackRate: 1,
    canGoNext: true,
    canGoPrevious: true,
    ...overrides,
});

describe('features/videos/playerShortcuts/playerShortcutExecution', () => {
    let controls: jest.Mocked<PlayerShortcutControls>;
    let muteMemory: { volumeBeforeMute: number };

    beforeEach(() => {
        controls = buildControls();
        muteMemory = { volumeBeforeMute: 0.5 };
    });

    const run = (
        action: Parameters<typeof runPlayerShortcut>[0],
        overrides: Partial<PlayerShortcutSnapshot> = {}
    ) => runPlayerShortcut(action, buildSnapshot(overrides), controls, muteMemory);

    it('clamps seek time between zero and the duration', () => {
        expect(clampSeekTime(-3, 100)).toBe(0);
        expect(clampSeekTime(130, 100)).toBe(100);
        expect(clampSeekTime(40, 100)).toBe(40);
        expect(clampSeekTime(40, 0)).toBe(40);
        expect(clampSeekTime(-1, Number.NaN)).toBe(0);
    });

    it('toggles play and pause with matching feedback', () => {
        expect(run({ type: 'togglePlayPause' })?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_PAUSE');
        expect(run({ type: 'togglePlayPause' }, { isPlaying: false })?.messageKey).toBe(
            'VIDEO_SHORTCUT_FEEDBACK_PLAY'
        );
        expect(controls.togglePlayPause).toHaveBeenCalledTimes(2);
    });

    it('seeks relatively within bounds', () => {
        expect(run({ type: 'seekRelative', seconds: 10 })).toEqual({
            messageKey: 'VIDEO_SHORTCUT_FEEDBACK_SEEK_FORWARD',
            params: { seconds: '10' },
        });
        expect(controls.seekTo).toHaveBeenLastCalledWith(60);

        run({ type: 'seekRelative', seconds: -5 }, { currentTime: 2 });
        expect(controls.seekTo).toHaveBeenLastCalledWith(0);

        const clampedForwardFeedback = run(
            { type: 'seekRelative', seconds: 10 },
            { currentTime: 95 }
        );
        expect(controls.seekTo).toHaveBeenLastCalledWith(100);
        expect(clampedForwardFeedback?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_SEEK_FORWARD');

        expect(run({ type: 'seekRelative', seconds: -10 })?.messageKey).toBe(
            'VIDEO_SHORTCUT_FEEDBACK_SEEK_BACKWARD'
        );
    });

    it('seeks to a fraction of the duration and ignores it without a known duration', () => {
        expect(run({ type: 'seekToFraction', fraction: 0.3 })).toEqual({
            messageKey: 'VIDEO_SHORTCUT_FEEDBACK_SEEK_PERCENT',
            params: { percent: '30' },
        });
        expect(controls.seekTo).toHaveBeenLastCalledWith(30);

        run({ type: 'seekToFraction', fraction: 1 });
        expect(controls.seekTo).toHaveBeenLastCalledWith(100);

        controls.seekTo.mockClear();
        expect(run({ type: 'seekToFraction', fraction: 0.3 }, { duration: 0 })).toBeNull();
        expect(run({ type: 'seekToFraction', fraction: 0.3 }, { duration: Number.NaN })).toBeNull();
        expect(controls.seekTo).not.toHaveBeenCalled();
    });

    it('adjusts volume within zero and one', () => {
        expect(run({ type: 'adjustVolume', delta: 0.05 })).toEqual({
            messageKey: 'VIDEO_SHORTCUT_FEEDBACK_VOLUME',
            params: { percent: '55' },
        });
        expect(controls.setVolume).toHaveBeenLastCalledWith(0.55);

        run({ type: 'adjustVolume', delta: 0.05 }, { volume: 0.98 });
        expect(controls.setVolume).toHaveBeenLastCalledWith(1);

        run({ type: 'adjustVolume', delta: -0.05 }, { volume: 0.02 });
        expect(controls.setVolume).toHaveBeenLastCalledWith(0);
    });

    it('mutes and restores the previous volume', () => {
        expect(run({ type: 'toggleMute' }, { volume: 0.8 })?.messageKey).toBe(
            'VIDEO_SHORTCUT_FEEDBACK_MUTED'
        );
        expect(controls.setVolume).toHaveBeenLastCalledWith(0);

        expect(run({ type: 'toggleMute' }, { volume: 0 })).toEqual({
            messageKey: 'VIDEO_SHORTCUT_FEEDBACK_VOLUME',
            params: { percent: '80' },
        });
        expect(controls.setVolume).toHaveBeenLastCalledWith(0.8);
    });

    it('restores the default volume when nothing was remembered', () => {
        muteMemory.volumeBeforeMute = 0;
        run({ type: 'toggleMute' }, { volume: 0 });
        expect(controls.setVolume).toHaveBeenLastCalledWith(0.5);
    });

    it('steps playback rate and clamps it to the supported range', () => {
        expect(run({ type: 'adjustPlaybackRate', delta: 0.25 })).toEqual({
            messageKey: 'VIDEO_SHORTCUT_FEEDBACK_SPEED',
            params: { rate: '1.25' },
        });
        run({ type: 'adjustPlaybackRate', delta: 0.25 }, { playbackRate: 2 });
        expect(controls.setPlaybackRate).toHaveBeenLastCalledWith(2);
        run({ type: 'adjustPlaybackRate', delta: -0.25 }, { playbackRate: 0.25 });
        expect(controls.setPlaybackRate).toHaveBeenLastCalledWith(0.25);
    });

    it('toggles fullscreen and opens help without feedback', () => {
        expect(run({ type: 'toggleFullscreen' })).toBeNull();
        expect(run({ type: 'openHelp' })).toBeNull();
        expect(controls.toggleFullscreen).toHaveBeenCalledTimes(1);
        expect(controls.openHelp).toHaveBeenCalledTimes(1);
    });

    it('navigates the playlist only when available', () => {
        expect(run({ type: 'nextVideo' })?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_NEXT');
        expect(run({ type: 'previousVideo' })?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_PREVIOUS');
        expect(controls.nextVideo).toHaveBeenCalledTimes(1);
        expect(controls.previousVideo).toHaveBeenCalledTimes(1);

        expect(run({ type: 'nextVideo' }, { canGoNext: false })).toBeNull();
        expect(run({ type: 'previousVideo' }, { canGoPrevious: false })).toBeNull();
        expect(controls.nextVideo).toHaveBeenCalledTimes(1);
        expect(controls.previousVideo).toHaveBeenCalledTimes(1);
    });
});
