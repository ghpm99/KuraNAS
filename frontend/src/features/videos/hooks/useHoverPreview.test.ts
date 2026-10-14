import { act, renderHook } from '@testing-library/react';
import { useHoverPreview } from './useHoverPreview';

const stubMatchMedia = (matchingQueries: string[]) => {
    window.matchMedia = jest.fn((query: string) => ({
        matches: matchingQueries.includes(query),
    })) as unknown as typeof window.matchMedia;
};

describe('useHoverPreview', () => {
    const originalMatchMedia = window.matchMedia;

    beforeEach(() => jest.useFakeTimers());

    afterEach(() => {
        jest.useRealTimers();
        window.matchMedia = originalMatchMedia;
    });

    it('stays unmounted and survives an environment without matchMedia', () => {
        window.matchMedia = undefined as unknown as typeof window.matchMedia;
        const { result } = renderHook(() => useHoverPreview());

        expect(result.current.isPreviewMounted).toBe(false);

        act(() => result.current.previewHandlers.onPointerEnter());
        act(() => {
            jest.advanceTimersByTime(300);
        });

        expect(result.current.isPreviewMounted).toBe(true);
    });

    it('mounts after the intent delay and not before', () => {
        const { result } = renderHook(() => useHoverPreview());

        act(() => result.current.previewHandlers.onPointerEnter());
        act(() => {
            jest.advanceTimersByTime(299);
        });
        expect(result.current.isPreviewMounted).toBe(false);

        act(() => {
            jest.advanceTimersByTime(1);
        });
        expect(result.current.isPreviewMounted).toBe(true);
    });

    it('cancels the request when the pointer leaves early', () => {
        const { result } = renderHook(() => useHoverPreview());

        act(() => result.current.previewHandlers.onPointerEnter());
        act(() => {
            jest.advanceTimersByTime(100);
        });
        act(() => result.current.previewHandlers.onPointerLeave());
        act(() => {
            jest.advanceTimersByTime(1000);
        });

        expect(result.current.isPreviewMounted).toBe(false);
    });

    it('mounts on focus, cancels on blur and stays mounted after leaving', () => {
        const { result } = renderHook(() => useHoverPreview());

        act(() => result.current.previewHandlers.onFocus());
        act(() => result.current.previewHandlers.onBlur());
        act(() => {
            jest.advanceTimersByTime(1000);
        });
        expect(result.current.isPreviewMounted).toBe(false);

        act(() => result.current.previewHandlers.onFocus());
        act(() => {
            jest.advanceTimersByTime(300);
        });
        act(() => result.current.previewHandlers.onBlur());
        act(() => result.current.previewHandlers.onPointerLeave());

        expect(result.current.isPreviewMounted).toBe(true);
    });

    it.each([['(hover: none)'], ['(prefers-reduced-motion: reduce)']])(
        'never mounts when %s matches',
        (query) => {
            stubMatchMedia([query]);
            const { result } = renderHook(() => useHoverPreview());

            act(() => result.current.previewHandlers.onPointerEnter());
            act(() => {
                jest.advanceTimersByTime(1000);
            });

            expect(result.current.isPreviewMounted).toBe(false);
        }
    );

    it('tracks load and unmounts for good after an error', () => {
        const { result } = renderHook(() => useHoverPreview());

        act(() => result.current.previewHandlers.onPointerEnter());
        act(() => {
            jest.advanceTimersByTime(300);
        });
        act(() => result.current.handlePreviewLoad());
        expect(result.current.isPreviewLoaded).toBe(true);

        act(() => result.current.handlePreviewError());
        expect(result.current.isPreviewMounted).toBe(false);
        expect(result.current.isPreviewLoaded).toBe(false);

        act(() => result.current.previewHandlers.onPointerEnter());
        act(() => {
            jest.advanceTimersByTime(1000);
        });
        expect(result.current.isPreviewMounted).toBe(false);
    });

    it('clears the pending timer on unmount', () => {
        const { result, unmount } = renderHook(() => useHoverPreview());

        act(() => result.current.previewHandlers.onPointerEnter());
        unmount();

        expect(jest.getTimerCount()).toBe(0);
    });
});
