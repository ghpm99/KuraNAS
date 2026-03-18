jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
    },
}));

import { apiBase } from './index';
import { applyUpdate, getUpdateStatus } from './update';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
};

describe('service/update', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it.each([
        {
            name: 'gets update status',
            fn: () => getUpdateStatus(),
            method: 'get' as const,
            url: '/update/status',
            response: { status: 'idle', available: false },
        },
        {
            name: 'applies update',
            fn: () => applyUpdate(),
            method: 'post' as const,
            url: '/update/apply',
            response: { started: true },
        },
    ])('$name', async ({ fn, method, url, response }) => {
        mockedApi[method].mockResolvedValue({ data: response });

        const result = await fn();

        expect(mockedApi[method]).toHaveBeenCalledWith(url);
        expect(result).toEqual(response);
    });
});
