import { act, fireEvent, renderHook } from '@testing-library/react';
import useControlsAutoHide from './useControlsAutoHide';

describe('videoControls/useControlsAutoHide', () => {
    beforeEach(() => {
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('starts visible', () => {
        const { result } = renderHook(() => useControlsAutoHide(false));

        expect(result.current.areControlsVisible).toBe(true);
    });

    it('hides after three idle seconds while playing', () => {
        const { result } = renderHook(() => useControlsAutoHide(true));

        act(() => {
            jest.advanceTimersByTime(2900);
        });
        expect(result.current.areControlsVisible).toBe(true);

        act(() => {
            jest.advanceTimersByTime(200);
        });
        expect(result.current.areControlsVisible).toBe(false);
    });

    it.each([
        ['mouse move', () => fireEvent.mouseMove(document)],
        ['key press', () => fireEvent.keyDown(document, { key: 'ArrowLeft' })],
        ['touch', () => fireEvent.touchStart(document)],
    ])('reappears on %s and hides again after inactivity', (_label, triggerActivity) => {
        const { result } = renderHook(() => useControlsAutoHide(true));
        act(() => {
            jest.advanceTimersByTime(3100);
        });
        expect(result.current.areControlsVisible).toBe(false);

        act(() => {
            triggerActivity();
        });
        expect(result.current.areControlsVisible).toBe(true);

        act(() => {
            jest.advanceTimersByTime(3100);
        });
        expect(result.current.areControlsVisible).toBe(false);
    });

    it('stays visible while paused and shows again when playback pauses', () => {
        const { result, rerender } = renderHook(({ isPlaying }) => useControlsAutoHide(isPlaying), {
            initialProps: { isPlaying: true },
        });
        act(() => {
            jest.advanceTimersByTime(3100);
        });
        expect(result.current.areControlsVisible).toBe(false);

        rerender({ isPlaying: false });
        expect(result.current.areControlsVisible).toBe(true);

        act(() => {
            jest.advanceTimersByTime(10000);
        });
        expect(result.current.areControlsVisible).toBe(true);
    });
});
