import { useQuery } from '@tanstack/react-query';
import { getAudioSummary } from '@/service/fileTypeMetadata';
import type { IMusicData } from '@/types/music';

interface TrackLyrics {
    lyrics: string;
    isLoading: boolean;
}

export const useTrackLyrics = (track: IMusicData | undefined, isEnabled: boolean): TrackLyrics => {
    const lyricsFromQueue = track?.metadata?.lyrics?.trim() ?? '';
    const shouldFetchLyrics = isEnabled && lyricsFromQueue === '' && track?.id !== undefined;

    const { data: audioSummary, isLoading } = useQuery({
        queryKey: ['music-lyrics', track?.id],
        queryFn: () => getAudioSummary(track!.id),
        enabled: shouldFetchLyrics,
        staleTime: Infinity,
        retry: false,
    });

    if (lyricsFromQueue !== '') return { lyrics: lyricsFromQueue, isLoading: false };
    return {
        lyrics: audioSummary?.lyrics?.trim() ?? '',
        isLoading: shouldFetchLyrics && isLoading,
    };
};
