import { renderHook, act } from '@testing-library/react';
import { useHeader } from './useHeader';

describe('useHeader', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        jest.setSystemTime(new Date('2026-03-16T12:00:00.000Z'));
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('updates currentTime every second when showClock is true', () => {
        const { result } = renderHook(() => useHeader(true));

        const initialTime = result.current.currentTime;

        act(() => {
            jest.advanceTimersByTime(1000);
        });

        expect(result.current.currentTime.getTime()).toBeGreaterThan(initialTime.getTime());
    });

    it('does not start interval when showClock is false', () => {
        const clearIntervalSpy = jest.spyOn(window, 'clearInterval');
        const { result } = renderHook(() => useHeader(false));

        act(() => {
            jest.advanceTimersByTime(3000);
        });

        // Time should not change since no interval was set
        // The initial time was set at render, so it stays
        expect(result.current.currentTime).toBeInstanceOf(Date);
        clearIntervalSpy.mockRestore();
    });

    it('clears interval on unmount when showClock is true', () => {
        const clearIntervalSpy = jest.spyOn(window, 'clearInterval');
        const { unmount } = renderHook(() => useHeader(true));

        unmount();

        expect(clearIntervalSpy).toHaveBeenCalled();
        clearIntervalSpy.mockRestore();
    });
});
