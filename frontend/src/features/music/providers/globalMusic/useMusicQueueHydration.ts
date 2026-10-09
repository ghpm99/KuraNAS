import { useEffect, useRef, useState } from 'react';
import type { IMusicData } from '@/types/music';
import { queueToTracks } from '@/features/music/components/musicQueueTracks';
import { getPlayerQueue, getPlayerState, type PlayerStateDto } from '@/service/playerState';
import { parseRepeatMode, type RepeatMode } from './repeatMode';

type HydrationCallbacks = {
    setQueue: (queue: IMusicData[]) => void;
    setCurrentIndex: (index: number) => void;
    setShuffle: (isShuffleEnabled: boolean) => void;
    setRepeatMode: (repeatMode: RepeatMode) => void;
    setVolume: (volume: number) => void;
    loadPausedTrack: (trackId: number, startPositionSeconds: number) => void;
};

const clampIndex = (index: number, queueLength: number) =>
    Math.min(Math.max(Number.isInteger(index) ? index : 0, 0), queueLength - 1);

const isValidVolume = (volume: number | undefined): volume is number =>
    typeof volume === 'number' && Number.isFinite(volume) && volume >= 0 && volume <= 1;

const fetchSavedPlayerState = (): Promise<PlayerStateDto | null> =>
    getPlayerState().catch(() => null);

export default function useMusicQueueHydration(enabled: boolean, callbacks: HydrationCallbacks) {
    const hasHydratedRef = useRef(false);
    const isMountedRef = useRef(true);
    const [hasSettled, setHasSettled] = useState(false);

    useEffect(() => {
        isMountedRef.current = true;
        return () => {
            isMountedRef.current = false;
        };
    }, []);

    useEffect(() => {
        if (!enabled || hasHydratedRef.current) {
            return;
        }
        hasHydratedRef.current = true;

        const restoreSavedPlayer = async () => {
            const [savedState, savedQueue] = await Promise.all([
                fetchSavedPlayerState(),
                getPlayerQueue(),
            ]);
            if (!isMountedRef.current) return;

            const tracks = queueToTracks(savedQueue);
            if (tracks.length === 0) return;

            const startIndex = clampIndex(savedQueue.current_index, tracks.length);
            const startTrack = tracks[startIndex]!;
            const isStateForStartTrack = savedState?.current_file_id === startTrack.id;

            callbacks.setQueue(tracks);
            callbacks.setCurrentIndex(startIndex);
            callbacks.loadPausedTrack(
                startTrack.id,
                isStateForStartTrack ? Math.max(savedState.current_position ?? 0, 0) : 0
            );
            if (!savedState) return;
            callbacks.setShuffle(Boolean(savedState.shuffle));
            callbacks.setRepeatMode(parseRepeatMode(savedState.repeat_mode));
            if (isValidVolume(savedState.volume)) {
                callbacks.setVolume(savedState.volume);
            }
        };

        restoreSavedPlayer()
            .catch(() => undefined)
            .finally(() => {
                if (isMountedRef.current) setHasSettled(true);
            });
    }, [enabled, callbacks]);

    return { hasSettled };
}
