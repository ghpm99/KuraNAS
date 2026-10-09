import {
    downloadOriginalUrl,
    FILMSTRIP_THUMBNAIL_SIZE,
    GRID_THUMBNAIL_SIZE,
    PREVIEW_IMAGE_SIZE,
    thumbnailSrcSet,
    thumbnailUrl,
    viewerImageUrl,
} from './imageThumbnailSources';

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => '/api/v1',
}));

describe('imageThumbnailSources', () => {
    it('requests a square box on the thumbnail endpoint', () => {
        expect(thumbnailUrl(7, GRID_THUMBNAIL_SIZE)).toBe(
            '/api/v1/files/thumbnail/7?width=400&height=400'
        );
        expect(thumbnailUrl(7, FILMSTRIP_THUMBNAIL_SIZE)).toBe(
            '/api/v1/files/thumbnail/7?width=160&height=160'
        );
    });

    it('offers 1x and 2x renditions through srcset', () => {
        expect(thumbnailSrcSet(7, GRID_THUMBNAIL_SIZE)).toBe(
            '/api/v1/files/thumbnail/7?width=400&height=400 400w, /api/v1/files/thumbnail/7?width=800&height=800 800w'
        );
    });

    it('serves the preview until the user zooms in, then the original', () => {
        expect(viewerImageUrl(7, 1)).toBe(
            `/api/v1/files/thumbnail/7?width=${PREVIEW_IMAGE_SIZE}&height=${PREVIEW_IMAGE_SIZE}`
        );
        expect(viewerImageUrl(7, 0.5)).toContain('/files/thumbnail/7');
        expect(viewerImageUrl(7, 1.2)).toBe('/api/v1/files/blob/7');
    });

    it('keeps formats browsers cannot render on the preview even when zoomed', () => {
        const previewUrl = `/api/v1/files/thumbnail/7?width=${PREVIEW_IMAGE_SIZE}&height=${PREVIEW_IMAGE_SIZE}`;
        for (const format of [
            '.heic',
            '.heif',
            '.tif',
            '.tiff',
            '.cr2',
            '.cr3',
            '.nef',
            '.arw',
            '.dng',
            '.orf',
            '.rw2',
            '.raf',
            '.srw',
            '.pef',
        ]) {
            expect(viewerImageUrl(7, 3, format)).toBe(previewUrl);
        }
        for (const format of ['.jpg', '.jfif', '.png', '.avif']) {
            expect(viewerImageUrl(7, 3, format)).toBe('/api/v1/files/blob/7');
        }
    });

    it('always points the download at the original file', () => {
        expect(downloadOriginalUrl(7)).toBe('/api/v1/files/download/7');
    });
});
