import { Box, Link } from '@mui/material';
import { Disc } from 'lucide-react';
import { Link as RouterLink } from 'react-router-dom';
import { getMusicRoute } from '@/app/routes';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import CoverArt from '@/features/music/components/CoverArt';
import { getArtistKeyFromLabel } from '@/features/music/components/contextMenu/musicGroupingKeys';
import { createAlbumPlaybackContext } from '@/features/music/components/playbackContext';
import { findStartIndex } from '@/features/music/components/musicQueueTracks';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/types/music';
import { getMusicAlbumSummary, getMusicByAlbum } from '@/service/music';
import { getAlbumCoverUrl } from '@/service/musicCover';
import { MusicAlbumSummary } from '@/types/music';
import { shuffleItems } from '@/utils/shuffleItems';
import { MUSIC_COLLECTION_PAGE_SIZE } from '../shared';
import { loadAlbumTracks } from '../libraryQueueLoaders';
import { useMusicGroupSummary } from '../useMusicGroupSummary';
import { useMusicInfinitePages } from '../useMusicInfinitePages';
import MusicCollectionTrackList from './MusicCollectionTrackList';

const ALBUM_COVER_SIZE = 220;
const ALBUM_COVER_REQUEST_SIZE = 512;

type AlbumPageProps = {
    albumKey: string;
    onBack: () => void;
};

const buildArtistPath = (artistName: string) =>
    `${getMusicRoute('artists')}?artist=${encodeURIComponent(getArtistKeyFromLabel(artistName))}`;

function AlbumSubtitle({ summary }: { summary: MusicAlbumSummary | null }) {
    if (!summary?.artist) {
        return null;
    }

    return (
        <>
            <Link component={RouterLink} to={buildArtistPath(summary.artist)} color="inherit">
                {summary.artist}
            </Link>
            {summary.year ? ` · ${summary.year}` : ''}
        </>
    );
}

export default function AlbumPage({ albumKey, onBack }: AlbumPageProps) {
    const { replaceQueue } = useGlobalMusic();
    const summary = useMusicGroupSummary<MusicAlbumSummary>(['music-album-summary', albumKey], () =>
        getMusicAlbumSummary(albumKey)
    );
    const tracksQuery = useMusicInfinitePages<IMusicData>(
        ['music-by-album', albumKey],
        (pageNumber) => getMusicByAlbum(albumKey, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracks = tracksQuery.items;
    const albumName = summary?.name ?? tracks[0]?.metadata?.album ?? albumKey;
    const playbackContext = createAlbumPlaybackContext(albumName);

    const queueAlbumTracks = async (trackId?: number, shouldShuffle = false) => {
        const allTracks = await loadAlbumTracks(albumKey);
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
                title={albumName}
                subtitle={<AlbumSubtitle summary={summary} />}
                trackCount={summary?.track_count}
                totalLengthSeconds={summary?.total_length_seconds}
                iconSize={ALBUM_COVER_SIZE}
                icon={
                    <CoverArt
                        src={getAlbumCoverUrl(albumKey, ALBUM_COVER_REQUEST_SIZE)}
                        fallback={<Disc size={64} opacity={0.7} />}
                    />
                }
                gradientFrom="#7c3aed"
                onBack={onBack}
                onPlayAll={() => void queueAlbumTracks()}
                onShuffleAll={() => void queueAlbumTracks(undefined, true)}
            />

            <MusicCollectionTrackList
                tracks={tracks}
                isAlbumLayout
                albumDiscCount={summary?.disc_count}
                isLoading={tracksQuery.isLoading}
                isError={tracksQuery.isError}
                errorMessage={tracksQuery.errorMessage}
                hasNextPage={tracksQuery.hasNextPage}
                isFetchingNextPage={tracksQuery.isFetchingNextPage}
                onPlayTrack={(track) => void queueAlbumTracks(track.id)}
                onRetry={tracksQuery.retry}
                fetchNextPage={tracksQuery.fetchNextPage}
            />
        </Box>
    );
}
