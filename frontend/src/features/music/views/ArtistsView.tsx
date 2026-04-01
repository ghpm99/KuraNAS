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
import { Play, User } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import CategoryHeader from '@/components/music/CategoryHeader';
import TrackListItem from '@/components/music/TrackListItem';
import { createArtistPlaybackContext } from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicArtists, getMusicByArtist } from '@/service/music';
import { MusicArtist } from '@/types/music';
import { Pagination } from '@/types/pagination';
import {
    handleKeyboardActivation,
    loadAllTracks,
    MUSIC_COLLECTION_PAGE_SIZE,
    shuffleTracks,
} from './shared';

const loadArtistTracks = (artistKey: string) =>
    loadAllTracks((page, pageSize) => getMusicByArtist(artistKey, page, pageSize));

export default function ArtistsView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedArtistKey = searchParams.get('artist') ?? '';
    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music-artists'],
        queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicArtist>> =>
            getMusicArtists(pageParam, MUSIC_COLLECTION_PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });
    const artists = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data]);
    const selectedArtist = useMemo(
        () => artists.find((artist) => artist.key === selectedArtistKey) ?? null,
        [artists, selectedArtistKey]
    );

    const handleSelectArtist = (artist: MusicArtist) => {
        setSearchParams((current) => {
            const next = new URLSearchParams(current);
            next.set('artist', artist.key);
            return next;
        });
    };

    const handleBack = () => {
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete('artist');
                return next;
            },
            { replace: true }
        );
    };

    if (selectedArtist) {
        return <ArtistTracksView artist={selectedArtist} onBack={handleBack} />;
    }

    return (
        <ArtistListView
            artists={artists}
            isLoading={isLoading}
            fetchNextPage={fetchNextPage}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onSelect={handleSelectArtist}
        />
    );
}

type ArtistListViewProps = {
    artists: MusicArtist[];
    isLoading: boolean;
    fetchNextPage: () => Promise<unknown>;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (artist: MusicArtist) => void;
};

function ArtistListView({
    artists,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: ArtistListViewProps) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();

    const handlePlayArtist = async (event: React.MouseEvent, artist: MusicArtist) => {
        event.stopPropagation();
        const tracks = await loadArtistTracks(artist.key);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createArtistPlaybackContext(artist.artist));
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
                {artists.map((artist) => (
                    <Grid key={artist.artist} size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}>
                        <Card
                            sx={{
                                bgcolor: 'background.paper',
                                transition: 'all 0.2s ease',
                                '&:hover': {
                                    bgcolor: 'rgba(255,255,255,0.04)',
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
                                onClick={() => onSelect(artist)}
                                onKeyDown={(event) =>
                                    handleKeyboardActivation(event, () => onSelect(artist))
                                }
                                sx={{ position: 'relative' }}
                            >
                                <Box
                                    sx={{
                                        pt: 2,
                                        display: 'flex',
                                        justifyContent: 'center',
                                    }}
                                >
                                    <Box
                                        sx={{
                                            width: 100,
                                            height: 100,
                                            borderRadius: '50%',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            bgcolor: 'primary.dark',
                                            boxShadow: '0 4px 16px rgba(0,0,0,0.3)',
                                        }}
                                    >
                                        <User size={40} opacity={0.7} />
                                    </Box>
                                </Box>
                                <CardContent
                                    sx={{
                                        p: 1.5,
                                        textAlign: 'center',
                                        '&:last-child': { pb: 1.5 },
                                    }}
                                >
                                    <Typography variant="subtitle2" fontWeight={600} noWrap>
                                        {artist.artist}
                                    </Typography>
                                    <Typography variant="caption" color="text.secondary">
                                        {artist.album_count} {t('MUSIC_ALBUMS')}
                                    </Typography>
                                </CardContent>
                                <IconButton
                                    className="play-overlay"
                                    onClick={(event) => void handlePlayArtist(event, artist)}
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

function ArtistTracksView({ artist, onBack }: { artist: MusicArtist; onBack: () => void }) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();
    const [menuAnchor, setMenuAnchor] = useState<{
        el: HTMLElement;
        fileId: number;
    } | null>(null);
    const playbackContext = createArtistPlaybackContext(artist.artist);

    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music-by-artist', artist.key],
        queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> =>
            getMusicByArtist(artist.key, pageParam, MUSIC_COLLECTION_PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const tracks = data?.pages.flatMap((page) => page.items) ?? [];

    const queueArtistTracks = async (trackId?: number, shuffle = false) => {
        const allTracks = await loadArtistTracks(artist.key);
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
                title={artist.artist}
                trackCount={tracks.length}
                icon={<User size={48} opacity={0.7} />}
                gradientFrom="#4f46e5"
                onBack={onBack}
                onPlayAll={() => void queueArtistTracks()}
                onShuffleAll={() => void queueArtistTracks(undefined, true)}
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
                            onPlay={(track) => void queueArtistTracks(track.id)}
                            onAddToPlaylist={(event, fileId) =>
                                setMenuAnchor({
                                    el: event.currentTarget as HTMLElement,
                                    fileId,
                                })
                            }
                            showArtist={false}
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
