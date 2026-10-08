jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import { apiBase } from './index';
import { getImageFiles } from './image';

const mockedApi = apiBase as unknown as { get: jest.Mock };

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
});
