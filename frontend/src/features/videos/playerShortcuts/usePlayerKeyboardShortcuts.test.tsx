import { act, fireEvent, render, renderHook } from '@testing-library/react';
import usePlayerKeyboardShortcuts, {
    SHORTCUT_FEEDBACK_DURATION_MS,
} from './usePlayerKeyboardShortcuts';
import type { PlayerShortcutControls, PlayerShortcutSnapshot } from './playerShortcutExecution';

const snapshot: PlayerShortcutSnapshot = {
    isPlaying: true,
    currentTime: 50,
    duration: 100,
    volume: 0.5,
    isMuted: false,
    playbackRate: 1,
    canGoNext: true,
    canGoPrevious: true,
};

const buildControls = (): jest.Mocked<PlayerShortcutControls> => ({
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    toggleMute: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePictureInPicture: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
    openHelp: jest.fn(),
    toggleSubtitles: jest.fn(),
});

const dispatchKey = (init: KeyboardEventInit, target: EventTarget = document.body) => {
    const event = new KeyboardEvent('keydown', { bubbles: true, cancelable: true, ...init });
    act(() => {
        target.dispatchEvent(event);
    });
    return event;
};

describe('features/videos/playerShortcuts/usePlayerKeyboardShortcuts', () => {
    beforeEach(() => {
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('mounts without shortcuts firing and with no feedback', () => {
        const controls = buildControls();
        const { result } = renderHook(() =>
            usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false })
        );

        expect(result.current.feedback).toBeNull();
        expect(controls.togglePlayPause).not.toHaveBeenCalled();
    });

    it('runs the mapped action and prevents the default for Space', () => {
        const controls = buildControls();
        const { result } = renderHook(() =>
            usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false })
        );

        const event = dispatchKey({ key: ' ' });

        expect(event.defaultPrevented).toBe(true);
        expect(controls.togglePlayPause).toHaveBeenCalledTimes(1);
        expect(result.current.feedback?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_PAUSE');
    });

    it('toggles subtitles with the C key and shows the feedback', () => {
        const controls = buildControls();
        controls.toggleSubtitles.mockReturnValue('enabled');
        const { result } = renderHook(() =>
            usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false })
        );

        const event = dispatchKey({ key: 'c' });

        expect(event.defaultPrevented).toBe(true);
        expect(controls.toggleSubtitles).toHaveBeenCalledTimes(1);
        expect(result.current.feedback?.messageKey).toBe('VIDEO_SHORTCUT_FEEDBACK_SUBTITLES_ON');
    });

    it('clears the feedback after the display duration', () => {
        const controls = buildControls();
        const { result } = renderHook(() =>
            usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false })
        );

        dispatchKey({ key: 'm' });
        expect(result.current.feedback).not.toBeNull();

        act(() => {
            jest.advanceTimersByTime(SHORTCUT_FEEDBACK_DURATION_MS);
        });
        expect(result.current.feedback).toBeNull();
    });

    it('does not prevent the default for unhandled keys', () => {
        const controls = buildControls();
        renderHook(() => usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false }));

        const event = dispatchKey({ key: 'x' });

        expect(event.defaultPrevented).toBe(false);
    });

    it('ignores keys typed into an input and keys with browser modifiers', () => {
        const controls = buildControls();
        renderHook(() => usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false }));
        const input = document.createElement('input');
        document.body.appendChild(input);

        const typingEvent = dispatchKey({ key: ' ' }, input);
        const modifiedEvent = dispatchKey({ key: 'k', ctrlKey: true });

        expect(typingEvent.defaultPrevented).toBe(false);
        expect(modifiedEvent.defaultPrevented).toBe(false);
        expect(controls.togglePlayPause).not.toHaveBeenCalled();
        input.remove();
    });

    it('ignores keys while suspended', () => {
        const controls = buildControls();
        renderHook(() => usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: true }));

        dispatchKey({ key: 'k' });

        expect(controls.togglePlayPause).not.toHaveBeenCalled();
    });

    it('ignores keys while a dialog or menu is open in the document', () => {
        const controls = buildControls();
        renderHook(() => usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false }));
        const dialog = document.createElement('div');
        dialog.setAttribute('role', 'dialog');
        document.body.appendChild(dialog);

        dispatchKey({ key: 'k' });
        expect(controls.togglePlayPause).not.toHaveBeenCalled();

        dialog.remove();
        dispatchKey({ key: 'k' });
        expect(controls.togglePlayPause).toHaveBeenCalledTimes(1);
    });

    it('uses the latest snapshot without reattaching the listener', () => {
        const controls = buildControls();
        const addSpy = jest.spyOn(document, 'addEventListener');
        const { rerender } = renderHook(
            ({ currentTime }) =>
                usePlayerKeyboardShortcuts({
                    snapshot: { ...snapshot, currentTime },
                    controls,
                    isSuspended: false,
                }),
            { initialProps: { currentTime: 10 } }
        );
        const keydownRegistrations = () =>
            addSpy.mock.calls.filter(([eventName]) => eventName === 'keydown').length;
        const registrationsAfterMount = keydownRegistrations();

        rerender({ currentTime: 40 });
        dispatchKey({ key: 'ArrowRight' });

        expect(controls.seekTo).toHaveBeenCalledWith(45);
        expect(keydownRegistrations()).toBe(registrationsAfterMount);
        addSpy.mockRestore();
    });

    it('toggles picture in picture with the I key', () => {
        const controls = buildControls();
        render(<HookHost controls={controls} />);

        const event = dispatchKey({ key: 'i' });

        expect(event.defaultPrevented).toBe(true);
        expect(controls.togglePictureInPicture).toHaveBeenCalledTimes(1);
    });

    it('removes the listener on unmount', () => {
        const controls = buildControls();
        const { unmount } = render(<HookHost controls={controls} />);

        unmount();
        const event = dispatchKey({ key: ' ' });

        expect(event.defaultPrevented).toBe(false);
        expect(controls.togglePlayPause).not.toHaveBeenCalled();
    });

    it('does not fire a stale timer after unmount', () => {
        const controls = buildControls();
        const { unmount } = render(<HookHost controls={controls} />);
        dispatchKey({ key: 'm' });

        unmount();

        expect(() => jest.advanceTimersByTime(SHORTCUT_FEEDBACK_DURATION_MS)).not.toThrow();
        fireEvent.keyDown(document.body, { key: 'm' });
    });
});

function HookHost({ controls }: { controls: PlayerShortcutControls }) {
    usePlayerKeyboardShortcuts({ snapshot, controls, isSuspended: false });
    return null;
}
