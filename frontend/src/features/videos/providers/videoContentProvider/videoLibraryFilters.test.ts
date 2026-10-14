import {
    hasActiveVideoLibraryFilters,
    noVideoLibraryFilters,
    toggleVideoLibraryFormat,
} from './videoLibraryFilters';

describe('videoLibraryFilters', () => {
    it('reports no active filter for the empty filters', () => {
        expect(hasActiveVideoLibraryFilters(noVideoLibraryFilters)).toBe(false);
    });

    it('reports an active filter for each kind', () => {
        expect(hasActiveVideoLibraryFilters({ formats: ['.mp4'] })).toBe(true);
        expect(hasActiveVideoLibraryFilters({ formats: [], resolution: 'hd' })).toBe(true);
        expect(hasActiveVideoLibraryFilters({ formats: [], duration: 'short' })).toBe(true);
    });

    it('adds and removes a format while keeping the other filters', () => {
        const withMp4 = toggleVideoLibraryFormat({ formats: [], resolution: 'hd' }, '.mp4');
        expect(withMp4).toEqual({ formats: ['.mp4'], resolution: 'hd' });

        const withoutMp4 = toggleVideoLibraryFormat(withMp4, '.mp4');
        expect(withoutMp4).toEqual({ formats: [], resolution: 'hd' });
    });
});
