import { Box, IconButton } from '@mui/material';
import { Play, Shuffle } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { createAllTracksPlaybackContext } from '@/features/music/components/playbackContext';
import { findStartIndex } from '@/features/music/components/musicQueueTracks';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/types/music';
import { getMusic } from '@/service/music';
import { shuffleItems } from '@/utils/shuffleItems';
import MusicCollectionTrackList from './components/MusicCollectionTrackList';
import { MUSIC_COLLECTION_PAGE_SIZE } from './shared';
import { useMusicInfinitePages } from './useMusicInfinitePages';

const playbackContext = createAllTracksPlaybackContext();

export default function AllTracksView() {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();
    const tracksQuery = useMusicInfinitePages<IMusicData>(['music-library-tracks'], (pageNumber) =>
        getMusic(pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const loadedTracks = tracksQuery.items;

    const playLoadedTracks = (startTrackId?: number) =>
        replaceQueue(loadedTracks, findStartIndex(loadedTracks, startTrackId), playbackContext);

    const shuffleLoadedTracks = () => replaceQueue(shuffleItems(loadedTracks), 0, playbackContext);

    return (
        <Box sx={{ p: 2 }}>
            {loadedTracks.length > 0 && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}>
                    <IconButton
                        onClick={() => playLoadedTracks()}
                        aria-label={t('MUSIC_PLAY_LOADED')}
                        sx={{
                            bgcolor: 'primary.main',
                            color: 'white',
                            width: 40,
                            height: 40,
                            '&:hover': { bgcolor: 'primary.light', transform: 'scale(1.05)' },
                            transition: 'all 0.2s ease',
                        }}
                    >
                        <Play size={20} fill="white" />
                    </IconButton>
                    <IconButton
                        onClick={shuffleLoadedTracks}
                        aria-label={t('MUSIC_SHUFFLE_LOADED')}
                        sx={{ color: 'text.secondary', '&:hover': { color: 'text.primary' } }}
                    >
                        <Shuffle size={20} />
                    </IconButton>
                </Box>
            )}
            <MusicCollectionTrackList
                tracks={loadedTracks}
                isLoading={tracksQuery.isLoading}
                isError={tracksQuery.isError}
                errorMessage={tracksQuery.errorMessage}
                hasNextPage={tracksQuery.hasNextPage}
                isFetchingNextPage={tracksQuery.isFetchingNextPage}
                onPlayTrack={(track) => playLoadedTracks(track.id)}
                onRetry={tracksQuery.retry}
                fetchNextPage={tracksQuery.fetchNextPage}
            />
        </Box>
    );
}
