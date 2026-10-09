import { useCallback, useEffect, useRef } from 'react';
import { updatePlayerState, type UpdatePlayerStateRequest } from '@/service/playerState';
import { flushPlayerState } from '@/service/playerStateFlush';
import type { MusicPlaybackContext } from '@/features/music/components/playbackContext';

const SYNC_DEBOUNCE_MS = 2000;

type SyncOverrides = {
    fileId?: number | null;
    position?: number;
    vol?: number;
    playlistId?: number | null;
};

type SyncDeps = {
    getCurrentTrackId: () => number | undefined;
    getCurrentTime: () => number;
    volume: number;
    shuffle: boolean;
    repeatMode: string;
    playbackContext: MusicPlaybackContext | undefined;
};

export default function useMusicStateSync(deps: SyncDeps) {
    const syncTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const depsRef = useRef(deps);

    useEffect(() => {
        depsRef.current = deps;
    });

    const buildStateRequest = useCallback((overrides?: SyncOverrides): UpdatePlayerStateRequest => {
        const currentDeps = depsRef.current;
        return {
            playlist_id:
                overrides?.playlistId !== undefined
                    ? overrides.playlistId
                    : (currentDeps.playbackContext?.playlistId ?? null),
            current_file_id:
                overrides?.fileId !== undefined
                    ? overrides.fileId
                    : (currentDeps.getCurrentTrackId() ?? null),
            current_position:
                overrides?.position !== undefined
                    ? overrides.position
                    : currentDeps.getCurrentTime(),
            volume: overrides?.vol !== undefined ? overrides.vol : currentDeps.volume,
            shuffle: currentDeps.shuffle,
            repeat_mode: currentDeps.repeatMode,
        };
    }, []);

    const syncState = useCallback(
        (overrides?: SyncOverrides) => {
            if (syncTimeoutRef.current) {
                clearTimeout(syncTimeoutRef.current);
            }
            syncTimeoutRef.current = setTimeout(() => {
                updatePlayerState(buildStateRequest(overrides)).catch(() => {});
            }, SYNC_DEBOUNCE_MS);
        },
        [buildStateRequest]
    );

    useEffect(() => {
        const flushCurrentState = () => {
            if (depsRef.current.getCurrentTrackId() === undefined) return;
            if (syncTimeoutRef.current) clearTimeout(syncTimeoutRef.current);
            flushPlayerState(buildStateRequest());
        };

        window.addEventListener('pagehide', flushCurrentState);
        return () => window.removeEventListener('pagehide', flushCurrentState);
    }, [buildStateRequest]);

    return { syncState };
}
