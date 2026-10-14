import { act, renderHook } from '@testing-library/react';
import { autoplayNextPreferenceStorageKey } from './autoplayNextPreference';
import useUpNextAutoplay, { UP_NEXT_COUNTDOWN_SECONDS } from './useUpNextAutoplay';

type HookProps = {
    currentVideoId: number | null;
    hasNextVideo: boolean;
    isPlaying: boolean;
    defaultAutoplayEnabled: boolean;
};

const initialProps: HookProps = {
    currentVideoId: 1,
    hasNextVideo: true,
    isPlaying: false,
    defaultAutoplayEnabled: true,
};

const renderUpNext = (onPlayNext = jest.fn(), props: HookProps = initialProps) => {
    const hook = renderHook(
        (currentProps: HookProps) => useUpNextAutoplay({ ...currentProps, onPlayNext }),
        { initialProps: props }
    );
    return { ...hook, onPlayNext };
};

const advanceSeconds = (seconds: number) => {
    for (let elapsed = 0; elapsed < seconds; elapsed += 1) {
        act(() => {
            jest.advanceTimersByTime(1000);
        });
    }
};

describe('features/videos/upNext/useUpNextAutoplay', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        window.localStorage.clear();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('starts closed without any provider or service', () => {
        const { result } = renderUpNext();

        expect(result.current.isOpen).toBe(false);
        expect(result.current.isCounting).toBe(false);
        expect(result.current.secondsRemaining).toBe(UP_NEXT_COUNTDOWN_SECONDS);
    });

    it('counts down and plays the next video after 8 seconds', () => {
        const { result, onPlayNext } = renderUpNext();

        act(() => result.current.open());
        expect(result.current.isCounting).toBe(true);

        advanceSeconds(7);
        expect(result.current.secondsRemaining).toBe(1);
        expect(onPlayNext).not.toHaveBeenCalled();

        advanceSeconds(1);
        expect(onPlayNext).toHaveBeenCalledTimes(1);
        expect(result.current.isOpen).toBe(false);
    });

    it('plays immediately on playNow and never again after the countdown', () => {
        const { result, onPlayNext } = renderUpNext();
        act(() => result.current.open());

        act(() => result.current.playNow());
        advanceSeconds(20);

        expect(onPlayNext).toHaveBeenCalledTimes(1);
        expect(result.current.isOpen).toBe(false);
    });

    it('cancel closes the card and keeps the next video from playing', () => {
        const { result, onPlayNext } = renderUpNext();
        act(() => result.current.open());

        act(() => result.current.cancel());
        advanceSeconds(20);

        expect(result.current.isOpen).toBe(false);
        expect(onPlayNext).not.toHaveBeenCalled();
    });

    it('does not open when there is no next video', () => {
        const { result, onPlayNext } = renderUpNext(jest.fn(), {
            ...initialProps,
            hasNextVideo: false,
        });

        act(() => result.current.open());
        advanceSeconds(20);

        expect(result.current.isOpen).toBe(false);
        expect(onPlayNext).not.toHaveBeenCalled();
    });

    it('cancels when playback resumes', () => {
        const { result, rerender, onPlayNext } = renderUpNext();
        act(() => result.current.open());

        rerender({ ...initialProps, isPlaying: true });
        advanceSeconds(20);

        expect(result.current.isOpen).toBe(false);
        expect(onPlayNext).not.toHaveBeenCalled();
    });

    it('closes when the current video changes', () => {
        const { result, rerender, onPlayNext } = renderUpNext();
        act(() => result.current.open());

        rerender({ ...initialProps, currentVideoId: 2 });
        advanceSeconds(20);

        expect(result.current.isOpen).toBe(false);
        expect(onPlayNext).not.toHaveBeenCalled();
    });

    it('shows the card without counting when autoplay is off', () => {
        window.localStorage.setItem(autoplayNextPreferenceStorageKey, 'false');
        const { result, onPlayNext } = renderUpNext();

        act(() => result.current.open());
        advanceSeconds(20);

        expect(result.current.isOpen).toBe(true);
        expect(result.current.isCounting).toBe(false);
        expect(result.current.secondsRemaining).toBe(UP_NEXT_COUNTDOWN_SECONDS);
        expect(onPlayNext).not.toHaveBeenCalled();
    });

    it('uses the default when no preference was stored', () => {
        const { result } = renderUpNext(jest.fn(), {
            ...initialProps,
            defaultAutoplayEnabled: false,
        });

        expect(result.current.isAutoplayEnabled).toBe(false);
    });

    it('persists the preference and stops the running countdown when turned off', () => {
        const { result, onPlayNext } = renderUpNext();
        act(() => result.current.open());
        advanceSeconds(3);

        act(() => result.current.setAutoplayEnabled(false));
        advanceSeconds(20);

        expect(window.localStorage.getItem(autoplayNextPreferenceStorageKey)).toBe('false');
        expect(result.current.isOpen).toBe(true);
        expect(onPlayNext).not.toHaveBeenCalled();

        const remounted = renderUpNext();
        expect(remounted.result.current.isAutoplayEnabled).toBe(false);
    });

    it('clears the pending timer on unmount', () => {
        const { result, unmount, onPlayNext } = renderUpNext();
        act(() => result.current.open());

        unmount();
        advanceSeconds(20);

        expect(jest.getTimerCount()).toBe(0);
        expect(onPlayNext).not.toHaveBeenCalled();
    });
});
