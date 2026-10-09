import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    Grid,
    IconButton,
    Typography,
} from '@mui/material';
import { Play, Tag } from 'lucide-react';
import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import { createGenrePlaybackContext } from '@/features/music/components/playbackContext';
import { queueToTracks, findStartIndex } from '@/features/music/components/musicQueueTracks';
import { shuffleItems } from '@/utils/shuffleItems';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicByGenre, getMusicGenres, getMusicQueueByGenre } from '@/service/music';
import { MusicGenre } from '@/types/music';
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

const GENRE_COLORS = [
    '#e11d48',
    '#db2777',
    '#c026d3',
    '#9333ea',
    '#7c3aed',
    '#6366f1',
    '#4f46e5',
    '#2563eb',
    '#0891b2',
    '#059669',
    '#d97706',
    '#ea580c',
    '#dc2626',
    '#be185d',
];

const getGenreColor = (genre: string) => {
    let hash = 0;
    for (let index = 0; index < genre.length; index += 1) {
        hash = genre.charCodeAt(index) + ((hash << 5) - hash);
    }
    return GENRE_COLORS[Math.abs(hash) % GENRE_COLORS.length];
};

const loadGenreTracks = (genreKey: string) => getMusicQueueByGenre(genreKey).then(queueToTracks);

export default function GenresView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedGenreKey = searchParams.get('genre') ?? '';
    const { listSort, changeField, toggleOrder } = useMusicListSort('genres');
    const genresQuery = useMusicInfinitePages<MusicGenre>(
        ['music-genres', listSort],
        (pageNumber) => getMusicGenres(pageNumber, MUSIC_COLLECTION_PAGE_SIZE, listSort)
    );
    const genres = genresQuery.items;
    const knownGenre = useMemo(
        () => genres.find((genre) => genre.key === selectedGenreKey) ?? null,
        [genres, selectedGenreKey]
    );

    const handleSelectGenre = (genre: MusicGenre) => {
        setSearchParams((current) => {
            const next = new URLSearchParams(current);
            next.set('genre', genre.key);
            return next;
        });
    };

    const handleBack = () => {
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete('genre');
                return next;
            },
            { replace: true }
        );
    };

    if (selectedGenreKey) {
        return (
            <GenreTracksView
                genreKey={selectedGenreKey}
                knownGenre={knownGenre}
                onBack={handleBack}
            />
        );
    }

    return (
        <>
            <MusicSortControl
                view="genres"
                listSort={listSort}
                onFieldChange={changeField}
                onOrderToggle={toggleOrder}
            />
            <GenreListView
                genres={genres}
                isLoading={genresQuery.isLoading}
                isError={genresQuery.isError}
                errorMessage={genresQuery.errorMessage}
                onRetry={genresQuery.retry}
                fetchNextPage={genresQuery.fetchNextPage}
                hasNextPage={genresQuery.hasNextPage}
                isFetchingNextPage={genresQuery.isFetchingNextPage}
                onSelect={handleSelectGenre}
            />
        </>
    );
}

type GenreListViewProps = {
    genres: MusicGenre[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    onRetry: () => void;
    fetchNextPage: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (genre: MusicGenre) => void;
};

function GenreListView({
    genres,
    isLoading,
    isError,
    errorMessage,
    onRetry,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: GenreListViewProps) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();

    const handlePlayGenre = async (event: React.MouseEvent, genre: MusicGenre) => {
        event.stopPropagation();
        const tracks = await loadGenreTracks(genre.key);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createGenrePlaybackContext(genre.genre));
        }
    };

    return (
        <MusicCollectionListFeedback
            isLoading={isLoading}
            isError={isError}
            errorMessage={errorMessage}
            isEmpty={genres.length === 0}
            emptyTitleKey="MUSIC_GENRES_EMPTY"
            errorTitleKey="MUSIC_LIST_ERROR_TITLE"
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onRetry={onRetry}
            fetchNextPage={fetchNextPage}
        >
            <Box sx={{ p: 2 }}>
                <Grid container spacing={2}>
                    {genres.map((genre) => {
                        const color = getGenreColor(genre.genre);
                        return (
                            <Grid key={genre.genre} size={{ xs: 6, sm: 4, md: 3 }}>
                                <Card
                                    sx={{
                                        bgcolor: 'background.paper',
                                        transition: 'all 0.2s ease',
                                        overflow: 'hidden',
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
                                        onClick={() => onSelect(genre)}
                                        onKeyDown={(event) =>
                                            handleKeyboardActivation(event, () => onSelect(genre))
                                        }
                                        sx={{ position: 'relative' }}
                                    >
                                        <Box
                                            sx={{
                                                height: 90,
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                background: `linear-gradient(135deg, ${color}cc 0%, ${color}66 100%)`,
                                                position: 'relative',
                                            }}
                                        >
                                            <Tag size={32} opacity={0.4} />
                                        </Box>
                                        <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
                                            <Typography variant="subtitle2" fontWeight={600} noWrap>
                                                {genre.genre}
                                            </Typography>
                                            <Typography variant="caption" color="text.secondary">
                                                {genre.track_count} {t('MUSIC_TRACKS_COUNT')}
                                            </Typography>
                                        </CardContent>
                                        <IconButton
                                            className="play-overlay"
                                            onClick={(event) => void handlePlayGenre(event, genre)}
                                            sx={{
                                                position: 'absolute',
                                                bottom: 42,
                                                right: 8,
                                                bgcolor: 'primary.main',
                                                color: 'white',
                                                width: 34,
                                                height: 34,
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
                                            <Play size={14} fill="white" />
                                        </IconButton>
                                    </CardActionArea>
                                </Card>
                            </Grid>
                        );
                    })}
                </Grid>
            </Box>
        </MusicCollectionListFeedback>
    );
}

function GenreTracksView({
    genreKey,
    knownGenre,
    onBack,
}: {
    genreKey: string;
    knownGenre: MusicGenre | null;
    onBack: () => void;
}) {
    const { replaceQueue } = useGlobalMusic();
    const tracksQuery = useMusicInfinitePages<IMusicData>(
        ['music-by-genre', genreKey],
        (pageNumber) => getMusicByGenre(genreKey, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracks = tracksQuery.items;
    const genreName = knownGenre?.genre ?? tracks[0]?.metadata?.genre ?? genreKey;
    const playbackContext = createGenrePlaybackContext(genreName);

    const queueGenreTracks = async (trackId?: number, shuffle = false) => {
        const allTracks = await loadGenreTracks(genreKey);
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
                title={genreName}
                trackCount={resolveCollectionTrackCount(
                    knownGenre?.track_count,
                    tracks.length,
                    tracksQuery.isFullyLoaded
                )}
                icon={<Tag size={48} opacity={0.7} />}
                gradientFrom={getGenreColor(genreName)}
                onBack={onBack}
                onPlayAll={() => void queueGenreTracks()}
                onShuffleAll={() => void queueGenreTracks(undefined, true)}
            />

            <MusicCollectionTrackList
                tracks={tracks}
                isLoading={tracksQuery.isLoading}
                isError={tracksQuery.isError}
                errorMessage={tracksQuery.errorMessage}
                hasNextPage={tracksQuery.hasNextPage}
                isFetchingNextPage={tracksQuery.isFetchingNextPage}
                onPlayTrack={(track) => void queueGenreTracks(track.id)}
                onRetry={tracksQuery.retry}
                fetchNextPage={tracksQuery.fetchNextPage}
            />
        </Box>
    );
}
