import { apiBase } from '@/service';
import { captureThumbnailHref, deleteCapture, getCaptureById, getCaptures } from './captures';
import type { Capture } from '@/types/captures';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        delete: jest.fn(),
    },
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://api/api/v1',
}));

const mockedGet = apiBase.get as jest.Mock;
const mockedDelete = apiBase.delete as jest.Mock;

const baseCapture = (overrides: Partial<Capture> = {}): Capture => ({
    id: 1,
    name: 'recording_1',
    file_name: 'recording_1.mp4',
    file_path: '/x/recording_1.mp4',
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 1024,
    episode_key: '',
    created_at: '2026-06-01T00:00:00Z',
    status: 'uploaded',
    ...overrides,
});

describe('captures service', () => {
    beforeEach(() => jest.clearAllMocks());

    it('getCaptures sends pagination and filters', async () => {
        mockedGet.mockResolvedValue({ data: { items: [], pagination: {} } });
        await getCaptures({ page: 2, pageSize: 10, name: 'frieren', mediaType: 'video' });
        expect(mockedGet).toHaveBeenCalledWith('/captures', {
            params: { page: 2, page_size: 10, name: 'frieren', media_type: 'video' },
        });
    });

    it('getCaptures uses defaults when called with no args', async () => {
        mockedGet.mockResolvedValue({ data: { items: [], pagination: {} } });
        await getCaptures();
        expect(mockedGet).toHaveBeenCalledWith('/captures', {
            params: { page: 1, page_size: 50, name: undefined, media_type: undefined },
        });
    });

    it('getCaptureById hits the detail route', async () => {
        mockedGet.mockResolvedValue({ data: baseCapture() });
        const result = await getCaptureById(7);
        expect(mockedGet).toHaveBeenCalledWith('/captures/7');
        expect(result.id).toBe(1);
    });

    it('deleteCapture hits the delete route', async () => {
        mockedDelete.mockResolvedValue({});
        await deleteCapture(9);
        expect(mockedDelete).toHaveBeenCalledWith('/captures/9');
    });

    describe('captureThumbnailHref', () => {
        it('prefers the provenance thumbnail_url', () => {
            expect(captureThumbnailHref(baseCapture({ thumbnail_url: 'http://x/t.jpg' }))).toBe(
                'http://x/t.jpg'
            );
        });

        it('falls back to the video thumbnail when promoted with a file_id', () => {
            expect(
                captureThumbnailHref(baseCapture({ status: 'promoted', file_id: 42 }))
            ).toBe('http://api/api/v1/files/video-thumbnail/42?width=320&height=180');
        });

        it('returns empty string when nothing is available', () => {
            expect(captureThumbnailHref(baseCapture({ status: 'uploaded' }))).toBe('');
            expect(captureThumbnailHref(baseCapture({ status: 'promoted' }))).toBe('');
        });
    });
});
