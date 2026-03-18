jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getActivityDiarySummary,
    getActivityDiaryEntries,
    createActivityDiaryEntry,
    duplicateActivityDiaryEntry,
} from './activityDiary';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
};

describe('service/activityDiary', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it.each([
        {
            name: 'gets activity diary summary',
            fn: () => getActivityDiarySummary(),
            method: 'get' as const,
            url: '/diary/summary',
            payload: undefined,
            response: { totalEntries: 10, streak: 5 },
        },
        {
            name: 'gets activity diary entries',
            fn: () => getActivityDiaryEntries(),
            method: 'get' as const,
            url: '/diary/',
            payload: undefined,
            response: { items: [{ id: 1, name: 'Entry 1' }], total: 1 },
        },
        {
            name: 'creates activity diary entry',
            fn: () => createActivityDiaryEntry({ name: 'Workout', description: 'Leg day' }),
            method: 'post' as const,
            url: '/diary/',
            payload: { name: 'Workout', description: 'Leg day' },
            response: { id: 1, name: 'Workout', description: 'Leg day' },
        },
        {
            name: 'duplicates activity diary entry',
            fn: () => duplicateActivityDiaryEntry(1),
            method: 'post' as const,
            url: '/diary/copy',
            payload: { ID: 1 },
            response: { id: 2, name: 'Workout copy' },
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
