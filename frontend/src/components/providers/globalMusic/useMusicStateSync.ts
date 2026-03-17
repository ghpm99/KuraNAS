import { useCallback, useEffect, useRef } from 'react';
import { updatePlayerState } from '@/service/playerState';
import type { MusicPlaybackContext } from '@/components/music/playbackContext';

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

    const syncState = useCallback((overrides?: SyncOverrides) => {
        if (syncTimeoutRef.current) {
            clearTimeout(syncTimeoutRef.current);
        }
        syncTimeoutRef.current = setTimeout(() => {
            const d = depsRef.current;
            updatePlayerState({
                playlist_id:
                    overrides?.playlistId !== undefined
                        ? overrides.playlistId
                        : (d.playbackContext?.playlistId ?? null),
                current_file_id:
                    overrides?.fileId !== undefined
                        ? overrides.fileId
                        : (d.getCurrentTrackId() ?? null),
                current_position:
                    overrides?.position !== undefined ? overrides.position : d.getCurrentTime(),
                volume: overrides?.vol !== undefined ? overrides.vol : d.volume,
                shuffle: d.shuffle,
                repeat_mode: d.repeatMode,
            }).catch(() => {});
        }, SYNC_DEBOUNCE_MS);
    }, []);

    return { syncState };
}
