import { act, renderHook } from '@testing-library/react';
import useMusicStateSync from './useMusicStateSync';
import { updatePlayerState } from '@/service/playerState';
import { flushPlayerState } from '@/service/playerStateFlush';
import {
    createPlaylistPlaybackContext,
    type MusicPlaybackContext,
} from '@/features/music/components/playbackContext';

jest.mock('@/service/playerState', () => ({
    updatePlayerState: jest.fn(),
}));

jest.mock('@/service/playerStateFlush', () => ({
    flushPlayerState: jest.fn(),
}));

describe('useMusicStateSync', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        (updatePlayerState as jest.Mock).mockReset();
        (flushPlayerState as jest.Mock).mockReset();
        (updatePlayerState as jest.Mock).mockResolvedValue(undefined);
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    const createDeps = (
        overrides?: Partial<{
            getCurrentTrackId: () => number | undefined;
            getCurrentTime: () => number;
            volume: number;
            shuffle: boolean;
            repeatMode: string;
            playbackContext: MusicPlaybackContext | undefined;
        }>
    ) => ({
        getCurrentTrackId: () => 1,
        getCurrentTime: () => 5,
        volume: 0.6,
        shuffle: true,
        repeatMode: 'all',
        playbackContext: createPlaylistPlaybackContext({ id: 12, name: 'test' }),
        ...overrides,
    });

    const runSync = () => {
        act(() => {
            jest.runOnlyPendingTimers();
        });
    };

    it('debounces and sends the latest override payload', () => {
        const { result } = renderHook(() => useMusicStateSync(createDeps()));

        act(() => {
            result.current.syncState();
            result.current.syncState({
                fileId: 5,
                position: 20,
                vol: 0.4,
                playlistId: 99,
            });
        });

        runSync();

        expect(updatePlayerState).toHaveBeenCalledTimes(1);
        expect(updatePlayerState).toHaveBeenCalledWith({
            current_file_id: 5,
            current_position: 20,
            volume: 0.4,
            playlist_id: 99,
            shuffle: true,
            repeat_mode: 'all',
        });
    });

    it('uses defaults when no overrides provided', () => {
        const { result } = renderHook(() => useMusicStateSync(createDeps()));

        act(() => {
            result.current.syncState();
        });

        runSync();

        expect(updatePlayerState).toHaveBeenCalledWith({
            current_file_id: 1,
            current_position: 5,
            volume: 0.6,
            playlist_id: 12,
            shuffle: true,
            repeat_mode: 'all',
        });
    });

    it('flushes the current state on pagehide and cancels the pending debounce', () => {
        const { result } = renderHook(() => useMusicStateSync(createDeps()));
        act(() => {
            result.current.syncState();
        });

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
        });
        runSync();

        expect(flushPlayerState).toHaveBeenCalledWith({
            current_file_id: 1,
            current_position: 5,
            volume: 0.6,
            playlist_id: 12,
            shuffle: true,
            repeat_mode: 'all',
        });
        expect(updatePlayerState).not.toHaveBeenCalled();
    });

    it('does not flush on pagehide when nothing is loaded', () => {
        renderHook(() => useMusicStateSync(createDeps({ getCurrentTrackId: () => undefined })));

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
        });

        expect(flushPlayerState).not.toHaveBeenCalled();
    });

    it('stops listening to pagehide after unmount', () => {
        const { unmount } = renderHook(() => useMusicStateSync(createDeps()));
        unmount();

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
        });

        expect(flushPlayerState).not.toHaveBeenCalled();
    });
});
