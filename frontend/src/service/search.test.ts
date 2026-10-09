jest.mock('.', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import { apiBase } from '.';
import { searchGlobal, searchGlobalWithAI } from './search';

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

    it('requests the AI endpoint with query and limit', async () => {
        const response = { ...emptySearchResult, query: 'my trip', suggestion: 'tip' };
        mockedApiGet.mockResolvedValue({ data: response });

        const result = await searchGlobalWithAI('my trip', 4);

        expect(mockedApiGet).toHaveBeenCalledWith('/search/global/ai', {
            params: { q: 'my trip', limit: 4 },
        });
        expect(result).toEqual(response);
    });

    it('forwards the abort signal to the global search request', async () => {
        mockedApiGet.mockResolvedValue({ data: emptySearchResult });
        const { signal } = new AbortController();

        await searchGlobal('mix', 6, signal);

        expect(mockedApiGet).toHaveBeenCalledWith('/search/global', {
            params: { q: 'mix', limit: 6 },
            signal,
        });
    });

    it('forwards the abort signal to the AI search request', async () => {
        mockedApiGet.mockResolvedValue({ data: emptySearchResult });
        const { signal } = new AbortController();

        await searchGlobalWithAI('my trip', 6, signal);

        expect(mockedApiGet).toHaveBeenCalledWith('/search/global/ai', {
            params: { q: 'my trip', limit: 6 },
            signal,
        });
    });
});
