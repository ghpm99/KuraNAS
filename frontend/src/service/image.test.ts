jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import axios from 'axios';
import type { ImageLibraryFilters } from '@/types/imageLibrary';
import { apiBase } from './index';
import {
    getImageFiles,
    getImageLibraryCount,
    getImageLibraryFolders,
    getImageLibraryPage,
    getImageLibraryTimeline,
    getImageMetadataSummary,
} from './image';

const mockedApi = apiBase as unknown as { get: jest.Mock };

const noFilters: ImageLibraryFilters = {
    nameQuery: '',
    categories: [],
    isStarredOnly: false,
    formats: [],
    takenFrom: '',
    takenTo: '',
    folder: '',
};

const allFilters: ImageLibraryFilters = {
    nameQuery: 'beach',
    categories: ['capture', 'screenshot_app'],
    isStarredOnly: true,
    formats: ['jpg', 'png'],
    takenFrom: '2026-01-01',
    takenTo: '2026-02-01',
    folder: '/photos/trip',
};

const repeatedKeys = { indexes: null };

describe('service/image', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('gets image files grouped as requested', async () => {
        const payload = { items: [], total: 0 };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getImageFiles(1, 30, 'date');

        expect(mockedApi.get).toHaveBeenCalledWith('/files/images', {
            params: { page: 1, page_size: 30, group_by: 'date' },
        });
        expect(result).toEqual(payload);
    });

    it('omits empty filters and sends only ordering and page size by default', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], has_next: false } });

        await getImageLibraryPage({
            filters: noFilters,
            ordering: { sort: 'taken_at', order: 'desc' },
            pageSize: 60,
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/image/library', {
            params: {
                q: undefined,
                category: undefined,
                starred: undefined,
                format: undefined,
                taken_from: undefined,
                taken_to: undefined,
                folder: undefined,
                sort: 'taken_at',
                order: 'desc',
                page_size: 60,
                cursor: undefined,
                page: undefined,
                taken_before: undefined,
            },
            paramsSerializer: repeatedKeys,
        });
    });

    it('maps every filter to its server parameter and repeats list keys', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], has_next: false } });

        await getImageLibraryPage({
            filters: allFilters,
            ordering: { sort: 'taken_at', order: 'desc' },
            pageSize: 60,
            cursor: 'cursor-token',
            takenBefore: '2025-04-01',
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/image/library', {
            params: expect.objectContaining({
                q: 'beach',
                category: ['capture', 'screenshot_app'],
                starred: true,
                format: ['jpg', 'png'],
                taken_from: '2026-01-01',
                taken_to: '2026-02-01',
                folder: '/photos/trip',
                cursor: 'cursor-token',
                taken_before: '2025-04-01',
            }),
            paramsSerializer: repeatedKeys,
        });

        const [url, requestConfig] = mockedApi.get.mock.calls[0]!;
        const queryString = axios.getUri({ url, ...requestConfig });
        expect(queryString).toContain('category=capture&category=screenshot_app');
        expect(queryString).toContain('format=jpg&format=png');
        expect(queryString).not.toContain('[]');
    });

    it('requests numbered pages for non keyset orderings', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], has_next: true, page: 2 } });

        const page = await getImageLibraryPage({
            filters: noFilters,
            ordering: { sort: 'name', order: 'asc' },
            pageSize: 60,
            page: 2,
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/image/library', {
            params: expect.objectContaining({ sort: 'name', order: 'asc', page: 2 }),
            paramsSerializer: repeatedKeys,
        });
        expect(page.page).toBe(2);
    });

    it('reads the total from the count endpoint with the same filters', async () => {
        mockedApi.get.mockResolvedValue({ data: { total: 42 } });

        const total = await getImageLibraryCount(allFilters);

        expect(total).toBe(42);
        expect(mockedApi.get).toHaveBeenCalledWith('/image/library/count', {
            params: expect.objectContaining({
                q: 'beach',
                category: ['capture', 'screenshot_app'],
            }),
            paramsSerializer: repeatedKeys,
        });
    });

    it('reads timeline buckets and tolerates an empty payload', async () => {
        mockedApi.get.mockResolvedValueOnce({ data: [{ year: 2026, month: 3, count: 5 }] });
        mockedApi.get.mockResolvedValueOnce({ data: null });

        expect(await getImageLibraryTimeline(allFilters)).toEqual([
            { year: 2026, month: 3, count: 5 },
        ]);
        expect(await getImageLibraryTimeline(noFilters)).toEqual([]);
        expect(mockedApi.get).toHaveBeenCalledWith('/image/library/timeline', {
            params: expect.objectContaining({ starred: true }),
            paramsSerializer: repeatedKeys,
        });
    });

    it('reads the EXIF summary of one image', async () => {
        mockedApi.get.mockResolvedValue({ data: { make: 'Sony' } });

        const summary = await getImageMetadataSummary(9);

        expect(mockedApi.get).toHaveBeenCalledWith('/image/metadata/9');
        expect(summary).toEqual({ make: 'Sony' });
    });

    it('requests the folders under a parent with pagination and omits an empty parent', async () => {
        const payload = { items: [], pagination: { page: 2, page_size: 48, has_next: false } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const nested = await getImageLibraryFolders('/photos', 2, 48);
        await getImageLibraryFolders('', 1, 48);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/image/library/folders', {
            params: { parent: '/photos', page: 2, page_size: 48 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/image/library/folders', {
            params: { parent: undefined, page: 1, page_size: 48 },
        });
        expect(nested).toEqual(payload);
    });
});
