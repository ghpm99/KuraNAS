import { useEffect, useRef } from 'react';
import type { IMusicData } from '../musicProvider/musicProvider';
import type { MusicPlaybackContext } from '@/components/music/playbackContext';
import { createPlaylistPlaybackContext } from '@/components/music/playbackContext';
import { getPlayerState } from '@/service/playerState';
import { getNowPlayingPlaylist, getPlaylistTracks } from '@/service/playlist';

type HydrationCallbacks = {
    setQueue: (queue: IMusicData[]) => void;
    setCurrentIndex: (index: number) => void;
    setPlaybackContext: (context: MusicPlaybackContext) => void;
};

export default function useMusicQueueHydration(enabled: boolean, callbacks: HydrationCallbacks) {
    const hasHydratedRef = useRef(false);

    useEffect(() => {
        if (!enabled || hasHydratedRef.current) {
            return;
        }

        hasHydratedRef.current = true;
        let cancelled = false;

        const hydrateQueue = async () => {
            try {
                const [playerState, nowPlayingPlaylist] = await Promise.all([
                    getPlayerState(),
                    getNowPlayingPlaylist(),
                ]);

                if (!nowPlayingPlaylist.id || !playerState.current_file_id) {
                    return;
                }

                const playlistTracks = await getPlaylistTracks(
                    nowPlayingPlaylist.id,
                    1,
                    Math.max(nowPlayingPlaylist.track_count, 1)
                );
                if (cancelled) return;

                const hydratedQueue = playlistTracks.items.map((item) => item.file);
                const startIndex = hydratedQueue.findIndex(
                    (track) => track.id === playerState.current_file_id
                );
                if (hydratedQueue.length === 0 || startIndex < 0) return;

                callbacks.setQueue(hydratedQueue);
                callbacks.setCurrentIndex(startIndex);
                callbacks.setPlaybackContext(createPlaylistPlaybackContext(nowPlayingPlaylist));
            } catch {
                // best effort hydration
            }
        };

        void hydrateQueue();

        return () => {
            cancelled = true;
        };
    }, [enabled, callbacks]);
}
