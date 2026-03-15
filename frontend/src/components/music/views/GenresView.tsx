import { Box, Card, CardActionArea, CardContent, CircularProgress, Grid, IconButton, List, Typography } from '@mui/material';
import { Play, Tag } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import CategoryHeader from '@/components/music/CategoryHeader';
import TrackListItem from '@/components/music/TrackListItem';
import { createGenrePlaybackContext } from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicByGenre, getMusicGenres } from '@/service/music';
import { MusicGenre } from '@/types/music';
import { Pagination } from '@/types/pagination';
import {
	handleKeyboardActivation,
	loadAllTracks,
	MUSIC_COLLECTION_PAGE_SIZE,
	shuffleTracks,
} from './shared';

const GENRE_COLORS = [
	'#e11d48', '#db2777', '#c026d3', '#9333ea', '#7c3aed',
	'#6366f1', '#4f46e5', '#2563eb', '#0891b2', '#059669',
	'#d97706', '#ea580c', '#dc2626', '#be185d',
];

const getGenreColor = (genre: string) => {
	let hash = 0;
	for (let index = 0; index < genre.length; index += 1) {
		hash = genre.charCodeAt(index) + ((hash << 5) - hash);
	}
	return GENRE_COLORS[Math.abs(hash) % GENRE_COLORS.length];
};

const loadGenreTracks = (genreKey: string) =>
	loadAllTracks((page, pageSize) => getMusicByGenre(genreKey, page, pageSize));

export default function GenresView() {
	const [selectedGenre, setSelectedGenre] = useState<MusicGenre | null>(null);

	if (selectedGenre) {
		return <GenreTracksView genre={selectedGenre} onBack={() => setSelectedGenre(null)} />;
	}

	return <GenreListView onSelect={setSelectedGenre} />;
}

function GenreListView({ onSelect }: { onSelect: (genre: MusicGenre) => void }) {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-genres'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicGenre>> => getMusicGenres(pageParam, MUSIC_COLLECTION_PAGE_SIZE),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const genres = data?.pages.flatMap((page) => page.items) ?? [];

	const handlePlayGenre = async (event: React.MouseEvent, genre: MusicGenre) => {
		event.stopPropagation();
		const tracks = await loadGenreTracks(genre.key);
		if (tracks.length > 0) {
			replaceQueue(tracks, 0, createGenrePlaybackContext(genre.genre));
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
				{genres.map((genre) => {
					const color = getGenreColor(genre.genre);
					return (
						<Grid key={genre.genre} size={{ xs: 6, sm: 4, md: 3 }}>
							<Card
								sx={{
									bgcolor: 'background.paper',
									transition: 'all 0.2s ease',
									overflow: 'hidden',
									'&:hover': { bgcolor: 'rgba(255,255,255,0.04)' },
									'&:hover .play-overlay': { opacity: 1, transform: 'translateY(0)' },
								}}
							>
								<CardActionArea
									component='div'
									role='button'
									tabIndex={0}
									onClick={() => onSelect(genre)}
									onKeyDown={(event) => handleKeyboardActivation(event, () => onSelect(genre))}
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
										<Typography variant='subtitle2' fontWeight={600} noWrap>
											{genre.genre}
										</Typography>
										<Typography variant='caption' color='text.secondary'>
											{genre.track_count} {t('MUSIC_TRACKS_COUNT')}
										</Typography>
									</CardContent>
									<IconButton
										className='play-overlay'
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
											boxShadow: '0 4px 12px rgba(99,102,241,0.4)',
											'&:hover': { bgcolor: 'primary.light', transform: 'translateY(0) scale(1.05)' },
										}}
									>
										<Play size={14} fill='white' />
									</IconButton>
								</CardActionArea>
							</Card>
						</Grid>
					);
				})}
			</Grid>

			{hasNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
					<Typography
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main', '&:hover': { textDecoration: 'underline' } }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}

function GenreTracksView({ genre, onBack }: { genre: MusicGenre; onBack: () => void }) {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const color = getGenreColor(genre.genre);
	const playbackContext = createGenrePlaybackContext(genre.genre);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-genre', genre.key],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> =>
			getMusicByGenre(genre.key, pageParam, MUSIC_COLLECTION_PAGE_SIZE),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const tracks = data?.pages.flatMap((page) => page.items) ?? [];

	const queueGenreTracks = async (trackId?: number, shuffle = false) => {
		const allTracks = await loadGenreTracks(genre.key);
		if (allTracks.length === 0) {
			return;
		}

		if (shuffle) {
			replaceQueue(shuffleTracks(allTracks), 0, playbackContext);
			return;
		}

		const startIndex = trackId ? Math.max(allTracks.findIndex((item) => item.id === trackId), 0) : 0;
		replaceQueue(allTracks, startIndex, playbackContext);
	};

	return (
		<Box sx={{ p: 2 }}>
			<CategoryHeader
				title={genre.genre}
				trackCount={tracks.length}
				icon={<Tag size={48} opacity={0.7} />}
				gradientFrom={color}
				onBack={onBack}
				onPlayAll={() => void queueGenreTracks()}
				onShuffleAll={() => void queueGenreTracks(undefined, true)}
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
							onPlay={(track) => void queueGenreTracks(track.id)}
							onAddToPlaylist={(event, fileId) => setMenuAnchor({ el: event.currentTarget as HTMLElement, fileId })}
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
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main', '&:hover': { textDecoration: 'underline' } }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}
