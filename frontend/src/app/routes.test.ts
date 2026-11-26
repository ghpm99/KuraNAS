import {
    appRoutes,
    buildDocumentsListUrl,
    buildDocumentViewerUrl,
    buildFavoritesUrl,
    buildFilesUrl,
    buildImageFolderUrl,
    buildVideoFolderUrl,
    buildVideoPlayerUrl,
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
    it('builds the video player url from the file id', () => {
        expect(buildVideoPlayerUrl(42)).toBe('/video/42');
    });

    it('builds the image and video folder urls with the encoded folder filter', () => {
        expect(buildImageFolderUrl('/Imagens/takeout')).toBe(
            '/images/folders?folder=%2FImagens%2Ftakeout'
        );
        expect(buildVideoFolderUrl('/Videos/takeout')).toBe(
            '/videos/folders?folder=%2FVideos%2Ftakeout'
        );
        expect(buildImageFolderUrl('')).toBe('/images/folders');
        expect(buildVideoFolderUrl('')).toBe('/videos/folders');
    });

    it('matches video player routes', () => {
        expect(isVideoPlayerRoute('/video/10')).toBe(true);
        expect(isVideoPlayerRoute('/music')).toBe(false);
    });

    it('builds music section routes and detects nested music paths', () => {
        expect(getMusicRoute('home')).toBe(appRoutes.music);
        expect(getMusicRoute('genres')).toBe('/music/genres');
        expect(getMusicRoute('tracks')).toBe('/music/tracks');
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
        expect(getFileBrowserRootPath('/favorites/docs/a')).toBe('/favorites');
        expect(getFileBrowserRootPath('/files/docs')).toBe('/files');
    });

    it('builds browser urls under the given root with encoded segments', () => {
        expect(buildFavoritesUrl('')).toBe('/favorites');
        expect(buildFavoritesUrl('/docs/my folder')).toBe('/favorites/docs/my%20folder');
        expect(buildFilesUrl('/docs')).toBe('/files/docs');
    });
});

describe('document viewer routes', () => {
    it('builds the viewer url keeping the list url and the page', () => {
        expect(buildDocumentViewerUrl(7, '')).toBe('/documents/view/7');
        expect(buildDocumentViewerUrl(7, '?format=pdf&q=a b', 3)).toBe(
            '/documents/view/7?from=format%3Dpdf%26q%3Da+b&page=3'
        );
        expect(buildDocumentViewerUrl(7, '', 2)).toBe('/documents/view/7?page=2');
    });

    it('builds the list url the viewer goes back to', () => {
        expect(buildDocumentsListUrl('')).toBe('/documents');
        expect(buildDocumentsListUrl('format=pdf')).toBe('/documents?format=pdf');
    });
});
