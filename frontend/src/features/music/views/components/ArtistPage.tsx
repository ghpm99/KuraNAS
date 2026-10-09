import { Box, Grid, Typography } from '@mui/material';
import { User } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { getMusicRoute } from '@/app/routes';
import ErrorState from '@/components/errorState/errorState';
import useI18n from '@/components/i18n/provider/i18nContext';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import { createArtistPlaybackContext } from '@/features/music/components/playbackContext';
import { findStartIndex } from '@/features/music/components/musicQueueTracks';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { getMusicAlbumsByArtist, getMusicArtistSummary, getMusicByArtist } from '@/service/music';
import { MusicAlbum, MusicArtistSummary } from '@/types/music';
import { shuffleItems } from '@/utils/shuffleItems';
import { MUSIC_COLLECTION_PAGE_SIZE } from '../shared';
import { loadArtistTracks } from '../libraryQueueLoaders';
import { useMusicGroupSummary } from '../useMusicGroupSummary';
import { useMusicInfinitePages } from '../useMusicInfinitePages';
import AlbumCard from './AlbumCard';
import MusicCollectionTrackList from './MusicCollectionTrackList';

type ArtistPageProps = {
    artistKey: string;
    onBack: () => void;
};

export default function ArtistPage({ artistKey, onBack }: ArtistPageProps) {
    const { t } = useI18n();
    const navigate = useNavigate();
    const { replaceQueue } = useGlobalMusic();
    const summary = useMusicGroupSummary<MusicArtistSummary>(
        ['music-artist-summary', artistKey],
        () => getMusicArtistSummary(artistKey)
    );
    const albumsQuery = useMusicInfinitePages<MusicAlbum>(
        ['music-albums-by-artist', artistKey],
        (pageNumber) => getMusicAlbumsByArtist(artistKey, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracksQuery = useMusicInfinitePages<IMusicData>(
        ['music-by-artist', artistKey],
        (pageNumber) => getMusicByArtist(artistKey, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracks = tracksQuery.items;
    const albums = albumsQuery.items;
    const artistName = summary?.name ?? tracks[0]?.metadata?.artist ?? artistKey;
    const playbackContext = createArtistPlaybackContext(artistName);

    const openAlbum = (album: MusicAlbum) =>
        navigate(`${getMusicRoute('albums')}?album=${encodeURIComponent(album.key)}`);

    const queueArtistTracks = async (trackId?: number, shouldShuffle = false) => {
        const allTracks = await loadArtistTracks(artistKey);
        if (allTracks.length === 0) {
            return;
        }

        if (shouldShuffle) {
            replaceQueue(shuffleItems(allTracks), 0, playbackContext);
            return;
        }

        replaceQueue(allTracks, findStartIndex(allTracks, trackId), playbackContext);
    };

    return (
        <Box sx={{ p: 2 }}>
            <CategoryHeader
                title={artistName}
                subtitle={summary ? `${summary.album_count} ${t('MUSIC_ALBUMS')}` : undefined}
                trackCount={summary?.track_count}
                totalLengthSeconds={summary?.total_length_seconds}
                icon={<User size={48} opacity={0.7} />}
                gradientFrom="#4f46e5"
                onBack={onBack}
                onPlayAll={() => void queueArtistTracks()}
                onShuffleAll={() => void queueArtistTracks(undefined, true)}
            />

            {albums.length > 0 && (
                <Box component="section" sx={{ mb: 3 }}>
                    <Typography component="h2" variant="h6" fontWeight={600} sx={{ mb: 1.5 }}>
                        {t('MUSIC_ALBUMS')}
                    </Typography>
                    <Grid container spacing={2}>
                        {albums.map((album) => (
                            <Grid key={album.key} size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}>
                                <AlbumCard album={album} onSelect={openAlbum} />
                            </Grid>
                        ))}
                    </Grid>
                    {albumsQuery.isError && (
                        <ErrorState
                            title={t('MUSIC_LIST_ERROR_TITLE')}
                            backendMessage={albumsQuery.errorMessage}
                            onRetry={albumsQuery.retry}
                        />
                    )}
                    <LoadMoreSentinel
                        hasNextPage={albumsQuery.hasNextPage && !albumsQuery.isError}
                        isFetchingNextPage={albumsQuery.isFetchingNextPage}
                        fetchNextPage={albumsQuery.fetchNextPage}
                    />
                </Box>
            )}

            <Typography component="h2" variant="h6" fontWeight={600} sx={{ mb: 1 }}>
                {t('MUSIC_ARTIST_ALL_TRACKS')}
            </Typography>
            <MusicCollectionTrackList
                tracks={tracks}
                isLoading={tracksQuery.isLoading}
                isError={tracksQuery.isError}
                errorMessage={tracksQuery.errorMessage}
                hasNextPage={tracksQuery.hasNextPage}
                isFetchingNextPage={tracksQuery.isFetchingNextPage}
                showArtist={false}
                onPlayTrack={(track) => void queueArtistTracks(track.id)}
                onRetry={tracksQuery.retry}
                fetchNextPage={tracksQuery.fetchNextPage}
            />
        </Box>
    );
}
