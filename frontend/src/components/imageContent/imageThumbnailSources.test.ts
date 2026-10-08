import {
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
});
