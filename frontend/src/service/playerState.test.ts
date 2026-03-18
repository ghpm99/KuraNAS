jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
    },
}));

import { apiBase } from './index';
import { getPlayerState, updatePlayerState } from './playerState';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
};

describe('service/playerState', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it.each([
        {
            name: 'gets player state',
            fn: () => getPlayerState(),
            method: 'get' as const,
            url: '/music/player-state/',
            payload: undefined,
            response: { id: 1, client_id: 'c1' },
        },
        {
            name: 'updates player state',
            fn: () => updatePlayerState({ current_position: 10, volume: 0.7 }),
            method: 'put' as const,
            url: '/music/player-state/',
            payload: { current_position: 10, volume: 0.7 },
            response: { id: 1, current_position: 10, volume: 0.7 },
        },
    ])('$name', async ({ fn, method, url, payload, response }) => {
        mockedApi[method].mockResolvedValue({ data: response });

        const result = await fn();

        if (payload) {
            expect(mockedApi[method]).toHaveBeenCalledWith(url, payload);
        } else {
            expect(mockedApi[method]).toHaveBeenCalledWith(url);
        }
        expect(result).toEqual(response);
    });
});
