import { useQuery } from '@tanstack/react-query';
import { useOptionalGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getMostPlayedTracks, getRecentlyPlayedTracks } from '@/service/music';
import type { MusicPlayedTrack } from '@/types/music';

const playHistoryRowSize = 6;

const playableTracksOf = (playedTracks: MusicPlayedTrack[] | undefined) =>
    (playedTracks ?? []).filter((playedTrack) => Boolean(playedTrack?.track));

export const useMusicPlayHistory = () => {
    const replaceQueue = useOptionalGlobalMusic()?.replaceQueue;

    const { data: mostPlayedPage, isLoading: isLoadingMostPlayed } = useQuery({
        queryKey: ['music-plays', 'most-played'],
        queryFn: () => getMostPlayedTracks(1, playHistoryRowSize, 'all'),
    });
    const { data: recentPlaysPage, isLoading: isLoadingRecentPlays } = useQuery({
        queryKey: ['music-plays', 'recent'],
        queryFn: () => getRecentlyPlayedTracks(1, playHistoryRowSize),
    });

    const mostPlayedTracks = playableTracksOf(mostPlayedPage?.items);
    const recentlyPlayedTracks = playableTracksOf(recentPlaysPage?.items);

    const playFromRow = (rowTracks: MusicPlayedTrack[], startIndex: number) =>
        replaceQueue?.(
            rowTracks.map((playedTrack) => playedTrack.track),
            startIndex
        );

    return {
        mostPlayedTracks,
        recentlyPlayedTracks,
        isLoadingMostPlayed,
        isLoadingRecentPlays,
        playMostPlayed: (startIndex: number) => playFromRow(mostPlayedTracks, startIndex),
        playRecentlyPlayed: (startIndex: number) => playFromRow(recentlyPlayedTracks, startIndex),
    };
};
