import { act, renderHook } from '@testing-library/react';
import useMusicQueuePersistence from './useMusicQueuePersistence';
import { apiBase } from '@/service';
import type { IMusicData } from '../musicProvider/musicProvider';

jest.mock('@/service', () => ({
    apiBase: { put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { put: jest.Mock };

describe('features/music/useMusicQueuePersistence (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
        mockedApi.put.mockResolvedValue({ data: {} });
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('debounced persistence issues PUT /music/player-state/queue with the body the backend decodes', () => {
        const { rerender } = renderHook(
            (props: { queue: IMusicData[]; currentIndex: number | undefined }) =>
                useMusicQueuePersistence({ isEnabled: true, ...props }),
            {
                initialProps: {
                    queue: [] as IMusicData[],
                    currentIndex: undefined as number | undefined,
                },
            }
        );

        rerender({ queue: [{ id: 3 }, { id: 8 }] as IMusicData[], currentIndex: 1 });
        act(() => {
            jest.advanceTimersByTime(2000);
        });

        expect(mockedApi.put).toHaveBeenCalledWith('/music/player-state/queue', {
            file_ids: [3, 8],
            current_index: 1,
        });
    });
});
