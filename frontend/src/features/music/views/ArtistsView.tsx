import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    Grid,
    IconButton,
    Typography,
} from '@mui/material';
import { Play, User } from 'lucide-react';
import { useSearchParams } from 'react-router-dom';
import CollectionContextMenu from '@/features/music/components/contextMenu/CollectionContextMenu';
import { createArtistPlaybackContext } from '@/features/music/components/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicArtists } from '@/service/music';
import { MusicArtist } from '@/types/music';
import { handleKeyboardActivation, MUSIC_COLLECTION_PAGE_SIZE } from './shared';
import ArtistPage from './components/ArtistPage';
import MusicCollectionListFeedback from './components/MusicCollectionListFeedback';
import MusicSortControl from './components/MusicSortControl';
import { useMusicInfinitePages } from './useMusicInfinitePages';
import { useMusicListSort } from './useMusicListSort';

import { loadArtistTracks } from './libraryQueueLoaders';

export default function ArtistsView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedArtistKey = searchParams.get('artist') ?? '';
    const { listSort, changeField, toggleOrder } = useMusicListSort('artists');
    const artistsQuery = useMusicInfinitePages<MusicArtist>(
        ['music-artists', listSort],
        (pageNumber) => getMusicArtists(pageNumber, MUSIC_COLLECTION_PAGE_SIZE, listSort)
    );
    const artists = artistsQuery.items;

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

    if (selectedArtistKey) {
        return <ArtistPage artistKey={selectedArtistKey} onBack={handleBack} />;
    }

    return (
        <>
            <MusicSortControl
                view="artists"
                listSort={listSort}
                onFieldChange={changeField}
                onOrderToggle={toggleOrder}
            />
            <ArtistListView
                artists={artists}
                isLoading={artistsQuery.isLoading}
                isError={artistsQuery.isError}
                errorMessage={artistsQuery.errorMessage}
                onRetry={artistsQuery.retry}
                fetchNextPage={artistsQuery.fetchNextPage}
                hasNextPage={artistsQuery.hasNextPage}
                isFetchingNextPage={artistsQuery.isFetchingNextPage}
                onSelect={handleSelectArtist}
            />
        </>
    );
}

type ArtistListViewProps = {
    artists: MusicArtist[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    onRetry: () => void;
    fetchNextPage: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (artist: MusicArtist) => void;
};

function ArtistListView({
    artists,
    isLoading,
    isError,
    errorMessage,
    onRetry,
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

    return (
        <MusicCollectionListFeedback
            isLoading={isLoading}
            isError={isError}
            errorMessage={errorMessage}
            isEmpty={artists.length === 0}
            emptyTitleKey="MUSIC_ARTISTS_EMPTY"
            errorTitleKey="MUSIC_LIST_ERROR_TITLE"
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onRetry={onRetry}
            fetchNextPage={fetchNextPage}
        >
            <Box sx={{ p: 2 }}>
                <Grid container spacing={2}>
                    {artists.map((artist) => (
                        <Grid key={artist.artist} size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}>
                            <CollectionContextMenu
                                collectionName={artist.artist}
                                playbackContext={createArtistPlaybackContext(artist.artist)}
                                loadTracks={() => loadArtistTracks(artist.key)}
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
                                            onClick={(event) =>
                                                void handlePlayArtist(event, artist)
                                            }
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
