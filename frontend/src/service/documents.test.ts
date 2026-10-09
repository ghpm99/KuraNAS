jest.mock('.', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import { apiBase } from '.';
import { searchDocuments } from './documents';

const mockedApiGet = apiBase.get as jest.Mock;

describe('service/documents', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('requests the document content search endpoint with the backend param names', async () => {
        const response = {
            items: [],
            pagination: { page: 2, page_size: 20, has_next: false, has_prev: true },
        };
        mockedApiGet.mockResolvedValue({ data: response });

        const result = await searchDocuments({ q: 'contrato', page: 2, pageSize: 20 });

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/search', {
            params: { q: 'contrato', page: 2, page_size: 20 },
            signal: undefined,
        });
        expect(result).toEqual(response);
    });
});
