import { act, renderHook } from '@testing-library/react';
import { resetFailedPosterUrls, useLocalPoster } from './useLocalPoster';

describe('useLocalPoster', () => {
    beforeEach(() => resetFailedPosterUrls());

    it('starts with the poster and switches to the fallback after one error', () => {
        const { result } = renderHook(() => useLocalPoster('/poster', '/thumb'));

        expect(result.current.imageUrl).toBe('/poster');
        expect(result.current.isPosterShown).toBe(true);

        act(() => result.current.handlePosterError());

        expect(result.current.imageUrl).toBe('/thumb');
        expect(result.current.isPosterShown).toBe(false);
    });

    it('remembers a failed poster so a remount does not request it again', () => {
        const first = renderHook(() => useLocalPoster('/poster', '/thumb'));
        act(() => first.result.current.handlePosterError());
        first.unmount();

        const second = renderHook(() => useLocalPoster('/poster', '/thumb'));

        expect(second.result.current.imageUrl).toBe('/thumb');
    });

    it('tries the new poster when the url changes after a failure', () => {
        const { result, rerender } = renderHook(
            ({ posterUrl }) => useLocalPoster(posterUrl, '/thumb'),
            { initialProps: { posterUrl: '/poster-a' } }
        );
        act(() => result.current.handlePosterError());

        rerender({ posterUrl: '/poster-b' });

        expect(result.current.imageUrl).toBe('/poster-b');
    });
});
