import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    CircularProgress,
    Grid,
    IconButton,
    List,
    Typography,
} from '@mui/material';
import { Disc, Play } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import CategoryHeader from '@/components/music/CategoryHeader';
import TrackListItem from '@/components/music/TrackListItem';
import { createAlbumPlaybackContext } from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicAlbums, getMusicByAlbum } from '@/service/music';
import { MusicAlbum } from '@/types/music';
import { Pagination } from '@/types/pagination';
import {
    handleKeyboardActivation,
    loadAllTracks,
    MUSIC_COLLECTION_PAGE_SIZE,
    shuffleTracks,
} from './shared';

const loadAlbumTracks = (albumKey: string) =>
    loadAllTracks((page, pageSize) => getMusicByAlbum(albumKey, page, pageSize));

export default function AlbumsView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedAlbumKey = searchParams.get('album') ?? '';
    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music-albums'],
        queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicAlbum>> =>
            getMusicAlbums(pageParam, MUSIC_COLLECTION_PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });
    const albums = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data]);
    const selectedAlbum = useMemo(
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

    if (selectedAlbum) {
        return <AlbumTracksView album={selectedAlbum} onBack={handleBack} />;
    }

    return (
        <AlbumListView
            albums={albums}
            isLoading={isLoading}
            fetchNextPage={fetchNextPage}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onSelect={handleSelectAlbum}
        />
    );
}

type AlbumListViewProps = {
    albums: MusicAlbum[];
    isLoading: boolean;
    fetchNextPage: () => Promise<unknown>;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (album: MusicAlbum) => void;
};

function AlbumListView({
    albums,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: AlbumListViewProps) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();

    const handlePlayAlbum = async (event: React.MouseEvent, album: MusicAlbum) => {
        event.stopPropagation();
        const tracks = await loadAlbumTracks(album.key);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createAlbumPlaybackContext(album.album));
        }
    };

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box sx={{ p: 2 }}>
            <Grid container spacing={2}>
                {albums.map((album) => (
                    <Grid
                        key={`${album.album}-${album.artist}`}
                        size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}
                    >
                        <Card
                            sx={{
                                bgcolor: 'background.paper',
                                transition: 'all 0.2s ease',
                                '&:hover': { bgcolor: 'rgba(255,255,255,0.04)' },
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
                                        boxShadow: '0 4px 12px rgba(99,102,241,0.4)',
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
                    </Grid>
                ))}
            </Grid>

            {hasNextPage && (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                    <Typography
                        variant="body2"
                        sx={{
                            cursor: 'pointer',
                            color: 'primary.main',
                            '&:hover': { textDecoration: 'underline' },
                        }}
                        onClick={() => fetchNextPage()}
                    >
                        {isFetchingNextPage ? (
                            <CircularProgress size={20} />
                        ) : (
                            t('ACTION_LOAD_MORE')
                        )}
                    </Typography>
                </Box>
            )}
        </Box>
    );
}

function AlbumTracksView({ album, onBack }: { album: MusicAlbum; onBack: () => void }) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();
    const [menuAnchor, setMenuAnchor] = useState<{
        el: HTMLElement;
        fileId: number;
    } | null>(null);
    const playbackContext = createAlbumPlaybackContext(album.album);

    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music-by-album', album.key],
        queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> =>
            getMusicByAlbum(album.key, pageParam, MUSIC_COLLECTION_PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const tracks = data?.pages.flatMap((page) => page.items) ?? [];

    const queueAlbumTracks = async (trackId?: number, shuffle = false) => {
        const allTracks = await loadAlbumTracks(album.key);
        if (allTracks.length === 0) {
            return;
        }

        if (shuffle) {
            replaceQueue(shuffleTracks(allTracks), 0, playbackContext);
            return;
        }

        const startIndex = trackId
            ? Math.max(
                  allTracks.findIndex((item) => item.id === trackId),
                  0
              )
            : 0;
        replaceQueue(allTracks, startIndex, playbackContext);
    };

    return (
        <Box sx={{ p: 2 }}>
            <CategoryHeader
                title={album.album}
                subtitle={album.artist}
                trackCount={tracks.length}
                icon={<Disc size={48} opacity={0.7} />}
                gradientFrom="#7c3aed"
                onBack={onBack}
                onPlayAll={() => void queueAlbumTracks()}
                onShuffleAll={() => void queueAlbumTracks(undefined, true)}
            />

            {isLoading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                    <CircularProgress />
                </Box>
            ) : (
                <List sx={{ width: '100%' }}>
                    {tracks.map((item, index) => (
                        <TrackListItem
                            key={item.id}
                            track={item}
                            index={index}
                            onPlay={(track) => void queueAlbumTracks(track.id)}
                            onAddToPlaylist={(event, fileId) =>
                                setMenuAnchor({
                                    el: event.currentTarget as HTMLElement,
                                    fileId,
                                })
                            }
                        />
                    ))}
                </List>
            )}

            <AddToPlaylistMenu
                fileId={menuAnchor?.fileId ?? 0}
                anchorEl={menuAnchor?.el ?? null}
                onClose={() => setMenuAnchor(null)}
            />

            {hasNextPage && (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                    <Typography
                        variant="body2"
                        sx={{
                            cursor: 'pointer',
                            color: 'primary.main',
                            '&:hover': { textDecoration: 'underline' },
                        }}
                        onClick={() => fetchNextPage()}
                    >
                        {isFetchingNextPage ? (
                            <CircularProgress size={20} />
                        ) : (
                            t('ACTION_LOAD_MORE')
                        )}
                    </Typography>
                </Box>
            )}
        </Box>
    );
}
