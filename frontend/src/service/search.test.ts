jest.mock('.', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import { apiBase } from '.';
import { searchGlobal } from './search';

const mockedApiGet = apiBase.get as jest.Mock;

const emptySearchResult = {
    files: [],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
};

describe('service/search', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it.each([
        {
            name: 'requests the global search endpoint with query and limit',
            fn: () => searchGlobal('mix', 8),
            params: { q: 'mix', limit: 8 },
            response: { ...emptySearchResult, query: 'mix' },
        },
        {
            name: 'uses the default per-section limit when omitted',
            fn: () => searchGlobal(''),
            params: { q: '', limit: 6 },
            response: { ...emptySearchResult, query: '' },
        },
    ])('$name', async ({ fn, params, response }) => {
        mockedApiGet.mockResolvedValue({ data: response });

        const result = await fn();

        expect(mockedApiGet).toHaveBeenCalledWith('/search/global', { params });
        expect(result).toEqual(response);
    });
});
