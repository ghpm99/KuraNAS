import { useCallback } from 'react';
import { useOptionalGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getMusicTrack } from '@/service/music';

/**
 * Returns a function that refetches only the track currently loaded in the player
 * and swaps it into the queue. Does nothing without a provider or a current track,
 * and ignores fetch failures (the player keeps the previous values).
 */
export default function useRefreshCurrentTrackMetadata(): () => Promise<void> {
    const globalMusic = useOptionalGlobalMusic();
    const currentTrackId = globalMusic?.currentTrack?.id;
    const refreshQueuedTracks = globalMusic?.refreshQueuedTracks;

    return useCallback(async () => {
        if (currentTrackId === undefined || !refreshQueuedTracks) return;
        try {
            refreshQueuedTracks([await getMusicTrack(currentTrackId)]);
        } catch {
            return;
        }
    }, [currentTrackId, refreshQueuedTracks]);
}
