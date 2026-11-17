jest.mock('./index', () => ({
    apiBase: { get: jest.fn() },
}));

import { apiBase } from './index';
import { getDuplicateGroupsPage } from './analytics';
import { buildServerPage } from '@/shared/test/buildServerPage';
import type { DuplicateGroupDetail } from '@/types/analytics';

const mockedApi = apiBase as unknown as { get: jest.Mock };

const duplicateGroup: DuplicateGroupDetail = {
    signature: 'abc',
    copies: 2,
    size_bytes: 10,
    reclaimable_size: 10,
    paths: ['/a/x.jpg', '/b/x.jpg'],
    files: [
        { id: 1, path: '/a/x.jpg', updated_at: '2026-01-01T00:00:00Z', tier: 'hot' },
        { id: 2, path: '/b/x.jpg', updated_at: '2026-01-02T00:00:00Z', tier: 'cold' },
    ],
};

describe('getDuplicateGroupsPage', () => {
    beforeEach(() => jest.clearAllMocks());

    it('sends page, page_size and the type filter and returns the envelope as received', async () => {
        const envelope = buildServerPage([duplicateGroup], {
            page: 3,
            pageSize: 50,
            hasNext: true,
        });
        mockedApi.get.mockResolvedValue({ data: envelope });

        const page = await getDuplicateGroupsPage({ page: 3, pageSize: 50, type: 'audio' });

        expect(mockedApi.get).toHaveBeenCalledWith('/analytics/duplicates/groups', {
            params: { page: 3, page_size: 50, type: 'audio' },
        });
        expect(page).toEqual(envelope);
    });

    it('omits type when no filter is chosen', async () => {
        mockedApi.get.mockResolvedValue({
            data: buildServerPage([], { page: 1, pageSize: 50, hasNext: false }),
        });

        await getDuplicateGroupsPage({ page: 1, pageSize: 50 });

        expect(mockedApi.get).toHaveBeenCalledWith('/analytics/duplicates/groups', {
            params: { page: 1, page_size: 50 },
        });
    });
});
