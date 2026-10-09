import { act, renderHook } from '@testing-library/react';
import useMusicQueueHydration from './useMusicQueueHydration';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

describe('features/music/useMusicQueueHydration (seam)', () => {
    it('reads GET /music/player-state/ and /music/player-state/queue shaped like the backend responds', async () => {
        mockedApi.get.mockImplementation((url: string) =>
            Promise.resolve({
                data:
                    url === '/music/player-state/queue'
                        ? {
                              items: [
                                  {
                                      file_id: 5,
                                      name: 'a.mp3',
                                      path: '/m/a.mp3',
                                      format: '.mp3',
                                      title: 'A',
                                      artist: 'B',
                                      album: 'C',
                                      length: 10,
                                  },
                              ],
                              current_index: 0,
                          }
                        : {
                              current_file_id: 5,
                              current_position: 4,
                              volume: 0.5,
                              shuffle: false,
                              repeat_mode: 'none',
                          },
            })
        );
        const callbacks = {
            setQueue: jest.fn(),
            setCurrentIndex: jest.fn(),
            setShuffle: jest.fn(),
            setRepeatMode: jest.fn(),
            setVolume: jest.fn(),
            loadPausedTrack: jest.fn(),
        };

        renderHook(() => useMusicQueueHydration(true, callbacks));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
            await Promise.resolve();
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/music/player-state/');
        expect(mockedApi.get).toHaveBeenCalledWith('/music/player-state/queue');
        expect(callbacks.loadPausedTrack).toHaveBeenCalledWith(5, 4);
    });
});
