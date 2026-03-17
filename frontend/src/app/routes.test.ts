import {
    appRoutes,
    getFileBrowserRootPath,
    getImageRoute,
    getMusicRoute,
    getVideoRoute,
    isImageRoute,
    isMusicRoute,
    isVideoPlayerRoute,
    isVideoRoute,
} from './routes';

describe('app routes helpers', () => {
    it('matches video player routes', () => {
        expect(isVideoPlayerRoute('/video/10')).toBe(true);
        expect(isVideoPlayerRoute('/music')).toBe(false);
    });

    it('builds music section routes and detects nested music paths', () => {
        expect(getMusicRoute('home')).toBe(appRoutes.music);
        expect(getMusicRoute('genres')).toBe('/music/genres');
        expect(isMusicRoute('/music')).toBe(true);
        expect(isMusicRoute('/music/playlists')).toBe(true);
        expect(isMusicRoute('/videos')).toBe(false);
    });

    it('builds image section routes and detects nested image paths', () => {
        expect(getImageRoute('library')).toBe(appRoutes.images);
        expect(getImageRoute('albums')).toBe('/images/albums');
        expect(isImageRoute('/images')).toBe(true);
        expect(isImageRoute('/images/folders')).toBe(true);
        expect(isImageRoute('/music')).toBe(false);
    });

    it('builds video section routes and detects nested video paths', () => {
        expect(getVideoRoute('home')).toBe(appRoutes.videos);
        expect(getVideoRoute('series')).toBe('/videos/series');
        expect(isVideoRoute('/videos')).toBe(true);
        expect(isVideoRoute('/videos/folders')).toBe(true);
        expect(isVideoRoute('/music')).toBe(false);
    });

    it('returns the correct browser root path', () => {
        expect(getFileBrowserRootPath('/files')).toBe('/files');
        expect(getFileBrowserRootPath('/favorites')).toBe('/favorites');
        expect(getFileBrowserRootPath('/starred')).toBe('/favorites');
    });
});
