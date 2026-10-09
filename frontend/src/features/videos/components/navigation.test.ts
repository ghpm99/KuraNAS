import {
    getVideoDetailRoute,
    getVideoDetailSlugFromPath,
    getVideoPlaylistIdFromSearch,
    getVideoSectionFromPath,
    getVideoSectionMeta,
    videoNavigationItems,
} from './navigation';

describe('video navigation helpers', () => {
    it('resolves the current section from known paths', () => {
        expect(getVideoSectionFromPath('/videos')).toBe('home');
        expect(getVideoSectionFromPath('/videos/clips')).toBe('clips');
        expect(getVideoSectionFromPath('/videos/series/breaking-bad')).toBe('series');
        expect(getVideoDetailSlugFromPath('/videos/series/breaking-bad')).toBe('breaking-bad');
        expect(getVideoDetailRoute('movies', 'the-matrix')).toBe('/videos/movies/the-matrix');
    });

    it('falls back to home metadata for unknown paths', () => {
        expect(getVideoSectionFromPath('/videos/unknown')).toBe('home');
        expect(getVideoSectionMeta('home')).toEqual(videoNavigationItems[0]);
    });

    it('appends the playlist id to the detail route when provided', () => {
        expect(getVideoDetailRoute('series', 'my show', 12)).toBe(
            '/videos/series/my%20show?playlist=12'
        );
        expect(getVideoDetailRoute('series', 'my show')).toBe('/videos/series/my%20show');
    });

    it('reads a valid playlist id from the search string', () => {
        expect(getVideoPlaylistIdFromSearch('?playlist=12')).toBe(12);
        expect(getVideoPlaylistIdFromSearch('?playlist=abc')).toBeNull();
        expect(getVideoPlaylistIdFromSearch('?playlist=-3')).toBeNull();
        expect(getVideoPlaylistIdFromSearch('')).toBeNull();
    });
});
