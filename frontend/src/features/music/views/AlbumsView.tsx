import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    Grid,
    IconButton,
    Typography,
} from '@mui/material';
import { Disc, Play } from 'lucide-react';
import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import CollectionContextMenu from '@/features/music/components/contextMenu/CollectionContextMenu';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import { createAlbumPlaybackContext } from '@/features/music/components/playbackContext';
import { queueToTracks, findStartIndex } from '@/features/music/components/musicQueueTracks';
import { shuffleItems } from '@/utils/shuffleItems';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { getMusicAlbums, getMusicByAlbum, getMusicQueueByAlbum } from '@/service/music';
import { MusicAlbum } from '@/types/music';
import {
    handleKeyboardActivation,
    MUSIC_COLLECTION_PAGE_SIZE,
    resolveCollectionTrackCount,
} from './shared';
import MusicCollectionListFeedback from './components/MusicCollectionListFeedback';
import MusicCollectionTrackList from './components/MusicCollectionTrackList';
import MusicSortControl from './components/MusicSortControl';
import { useMusicInfinitePages } from './useMusicInfinitePages';
import { useMusicListSort } from './useMusicListSort';

const loadAlbumTracks = (albumKey: string) => getMusicQueueByAlbum(albumKey).then(queueToTracks);

export default function AlbumsView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedAlbumKey = searchParams.get('album') ?? '';
    const { listSort, changeField, toggleOrder } = useMusicListSort('albums');
    const albumsQuery = useMusicInfinitePages<MusicAlbum>(
        ['music-albums', listSort],
        (pageNumber) => getMusicAlbums(pageNumber, MUSIC_COLLECTION_PAGE_SIZE, listSort)
    );
    const albums = albumsQuery.items;
    const knownAlbum = useMemo(
        () => albums.find((album) => album.key === selectedAlbumKey) ?? null,
        [albums, selectedAlbumKey]
    );

    const handleSelectAlbum = (album: MusicAlbum) => {
        setSearchParams((current) => {
            const next = new URLSearchParams(current);
            next.set('album', album.key);
            return next;
        });
    };

    const handleBack = () => {
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete('album');
                return next;
            },
            { replace: true }
        );
    };

    if (selectedAlbumKey) {
        return (
            <AlbumTracksView
                albumKey={selectedAlbumKey}
                knownAlbum={knownAlbum}
                onBack={handleBack}
            />
        );
    }

    return (
        <>
            <MusicSortControl
                view="albums"
                listSort={listSort}
                onFieldChange={changeField}
                onOrderToggle={toggleOrder}
            />
            <AlbumListView
                albums={albums}
                isLoading={albumsQuery.isLoading}
                isError={albumsQuery.isError}
                errorMessage={albumsQuery.errorMessage}
                onRetry={albumsQuery.retry}
                fetchNextPage={albumsQuery.fetchNextPage}
                hasNextPage={albumsQuery.hasNextPage}
                isFetchingNextPage={albumsQuery.isFetchingNextPage}
                onSelect={handleSelectAlbum}
            />
        </>
    );
}

type AlbumListViewProps = {
    albums: MusicAlbum[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    onRetry: () => void;
    fetchNextPage: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (album: MusicAlbum) => void;
};

function AlbumListView({
    albums,
    isLoading,
    isError,
    errorMessage,
    onRetry,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: AlbumListViewProps) {
    const { replaceQueue } = useGlobalMusic();

    const handlePlayAlbum = async (event: React.MouseEvent, album: MusicAlbum) => {
        event.stopPropagation();
        const tracks = await loadAlbumTracks(album.key);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createAlbumPlaybackContext(album.album));
        }
    };

    return (
        <MusicCollectionListFeedback
            isLoading={isLoading}
            isError={isError}
            errorMessage={errorMessage}
            isEmpty={albums.length === 0}
            emptyTitleKey="MUSIC_ALBUMS_EMPTY"
            errorTitleKey="MUSIC_LIST_ERROR_TITLE"
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onRetry={onRetry}
            fetchNextPage={fetchNextPage}
        >
            <Box sx={{ p: 2 }}>
                <Grid container spacing={2}>
                    {albums.map((album) => (
                        <Grid
                            key={`${album.album}-${album.artist}`}
                            size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}
                        >
                            <CollectionContextMenu
                                collectionName={album.album}
                                playbackContext={createAlbumPlaybackContext(album.album)}
                                loadTracks={() => loadAlbumTracks(album.key)}
                                layout="card"
                            >
                                <Card
                                    sx={{
                                        bgcolor: 'background.paper',
                                        transition: 'all 0.2s ease',
                                        '&:hover': {
                                            bgcolor: 'rgba(var(--app-color-ink-rgb), 0.04)',
                                        },
                                        '&:hover .play-overlay': {
                                            opacity: 1,
                                            transform: 'translateY(0)',
                                        },
                                    }}
                                >
                                    <CardActionArea
                                        component="div"
                                        role="button"
                                        tabIndex={0}
                                        onClick={() => onSelect(album)}
                                        onKeyDown={(event) =>
                                            handleKeyboardActivation(event, () => onSelect(album))
                                        }
                                        sx={{ position: 'relative' }}
                                    >
                                        <Box
                                            sx={{
                                                height: 140,
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                bgcolor: 'secondary.dark',
                                                background:
                                                    'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)',
                                            }}
                                        >
                                            <Disc size={48} opacity={0.5} />
                                        </Box>
                                        <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
                                            <Typography variant="subtitle2" fontWeight={600} noWrap>
                                                {album.album}
                                            </Typography>
                                            <Typography
                                                variant="caption"
                                                color="text.secondary"
                                                noWrap
                                                component="div"
                                            >
                                                {album.artist} {album.year ? `· ${album.year}` : ''}
                                            </Typography>
                                        </CardContent>
                                        <IconButton
                                            className="play-overlay"
                                            onClick={(event) => void handlePlayAlbum(event, album)}
                                            sx={{
                                                position: 'absolute',
                                                bottom: 50,
                                                right: 8,
                                                bgcolor: 'primary.main',
                                                color: 'white',
                                                width: 36,
                                                height: 36,
                                                opacity: 0,
                                                transform: 'translateY(8px)',
                                                transition: 'all 0.2s ease',
                                                boxShadow:
                                                    '0 4px 12px rgba(var(--app-color-primary-rgb), 0.4)',
                                                '&:hover': {
                                                    bgcolor: 'primary.light',
                                                    transform: 'translateY(0) scale(1.05)',
                                                },
                                            }}
                                        >
                                            <Play size={16} fill="white" />
                                        </IconButton>
                                    </CardActionArea>
                                </Card>
                            </CollectionContextMenu>
                        </Grid>
                    ))}
                </Grid>
            </Box>
        </MusicCollectionListFeedback>
    );
}

function AlbumTracksView({
    albumKey,
    knownAlbum,
    onBack,
}: {
    albumKey: string;
    knownAlbum: MusicAlbum | null;
    onBack: () => void;
}) {
    const { replaceQueue } = useGlobalMusic();
    const tracksQuery = useMusicInfinitePages<IMusicData>(
        ['music-by-album', albumKey],
        (pageNumber) => getMusicByAlbum(albumKey, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracks = tracksQuery.items;
    const albumName = knownAlbum?.album ?? tracks[0]?.metadata?.album ?? albumKey;
    const playbackContext = createAlbumPlaybackContext(albumName);

    const queueAlbumTracks = async (trackId?: number, shuffle = false) => {
        const allTracks = await loadAlbumTracks(albumKey);
        if (allTracks.length === 0) {
            return;
        }

        if (shuffle) {
            replaceQueue(shuffleItems(allTracks), 0, playbackContext);
            return;
        }

        const startIndex = findStartIndex(allTracks, trackId);
        replaceQueue(allTracks, startIndex, playbackContext);
    };

    return (
        <Box sx={{ p: 2 }}>
            <CategoryHeader
                title={albumName}
                subtitle={
                    knownAlbum?.artist ??
                    tracks[0]?.metadata?.album_artist ??
                    tracks[0]?.metadata?.artist
                }
                trackCount={resolveCollectionTrackCount(
                    knownAlbum?.track_count,
                    tracks.length,
                    tracksQuery.isFullyLoaded
                )}
                icon={<Disc size={48} opacity={0.7} />}
                gradientFrom="#7c3aed"
                onBack={onBack}
                onPlayAll={() => void queueAlbumTracks()}
                onShuffleAll={() => void queueAlbumTracks(undefined, true)}
            />

            <MusicCollectionTrackList
                tracks={tracks}
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
