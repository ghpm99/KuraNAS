jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        put: jest.fn(),
        delete: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    deleteTrashItem,
    emptyTrash,
    getTrashItems,
    getTrashRetention,
    getTrashSummary,
    restoreTrashItem,
    updateTrashRetention,
} from './trash';
import type { TrashPage } from '@/types/trash';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    put: jest.Mock;
    delete: jest.Mock;
};

const samplePage: TrashPage = {
    items: [
        {
            id: 1,
            original_path: '/data/docs/a.txt',
            size: 8,
            deleted_at: '2026-06-11T10:00:00Z',
        },
    ],
    pagination: { page: 1, page_size: 15, has_next: false, has_prev: false },
};

describe('service/trash', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('lists trash items with pagination and sorting params', async () => {
        mockedApi.get.mockResolvedValue({ data: samplePage });
        const result = await getTrashItems(2, 20, {
            sortKey: 'size',
            order: 'asc',
            searchText: '',
        });
        expect(mockedApi.get).toHaveBeenCalledWith('/trash', {
            params: { page: 2, page_size: 20, sort: 'size', order: 'asc' },
        });
        expect(result).toEqual(samplePage);
    });

    it('sends the trimmed search text only when there is one', async () => {
        mockedApi.get.mockResolvedValue({ data: samplePage });
        await getTrashItems(1, 15, { sortKey: 'name', order: 'desc', searchText: '  fotos ' });
        expect(mockedApi.get).toHaveBeenCalledWith('/trash', {
            params: { page: 1, page_size: 15, sort: 'name', order: 'desc', q: 'fotos' },
        });
    });

    it('restores an item by id', async () => {
        mockedApi.post.mockResolvedValue({ data: {} });
        await restoreTrashItem(7);
        expect(mockedApi.post).toHaveBeenCalledWith('/trash/7/restore');
    });

    it('permanently deletes one item', async () => {
        mockedApi.delete.mockResolvedValue({ data: {} });
        await deleteTrashItem(7);
        expect(mockedApi.delete).toHaveBeenCalledWith('/trash/7');
    });

    it('empties the trash', async () => {
        mockedApi.delete.mockResolvedValue({ data: {} });
        await emptyTrash();
        expect(mockedApi.delete).toHaveBeenCalledWith('/trash');
    });

    it('reads the retention policy', async () => {
        mockedApi.get.mockResolvedValue({ data: { days: 30 } });
        const result = await getTrashRetention();
        expect(mockedApi.get).toHaveBeenCalledWith('/trash/retention');
        expect(result).toEqual({ days: 30 });
    });

    it('updates the retention policy', async () => {
        mockedApi.put.mockResolvedValue({ data: { days: 7 } });
        const result = await updateTrashRetention(7);
        expect(mockedApi.put).toHaveBeenCalledWith('/trash/retention', { days: 7 });
        expect(result).toEqual({ days: 7 });
    });

    it('reads the trash totals', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: 3, size_bytes: 2048 } });

        const summary = await getTrashSummary();

        expect(mockedApi.get).toHaveBeenCalledWith('/trash/summary');
        expect(summary).toEqual({ items: 3, size_bytes: 2048 });
    });
});
