import { Box, Card, CardActionArea, CardContent, CircularProgress, Grid, IconButton, List, Typography } from '@mui/material';
import { Play, User } from 'lucide-react';
import { useState } from 'react';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import { createArtistPlaybackContext } from '@/components/music/playbackContext';
import TrackListItem from '@/components/music/TrackListItem';
import CategoryHeader from '@/components/music/CategoryHeader';
import { useInfiniteQuery } from '@tanstack/react-query';
import { getMusicArtists, getMusicByArtist } from '@/service/music';
import { MusicArtist } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';

const ArtistsView = () => {
	const [selectedArtist, setSelectedArtist] = useState<string | null>(null);

	if (selectedArtist) {
		return <ArtistTracksView artist={selectedArtist} onBack={() => setSelectedArtist(null)} />;
	}

	return <ArtistListView onSelect={setSelectedArtist} />;
};

const handleActionAreaKeyDown = (event: React.KeyboardEvent<HTMLElement>, onActivate: () => void) => {
	if (event.key === 'Enter' || event.key === ' ') {
		event.preventDefault();
		onActivate();
	}
};

const ArtistListView = ({ onSelect }: { onSelect: (artist: string) => void }) => {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-artists'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicArtist>> => {
			return getMusicArtists(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const artists = data?.pages.flatMap((page) => page.items) ?? [];

	const handlePlayArtist = async (e: React.MouseEvent, artist: string) => {
		e.stopPropagation();
		const data = await getMusicByArtist(artist, 1, 200);
		if (data.items.length > 0) replaceQueue(data.items, 0, createArtistPlaybackContext(artist));
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
								'&:hover .play-overlay': { opacity: 1, transform: 'translateY(0)' },
							}}
						>
							<CardActionArea
								component='div'
								role='button'
								tabIndex={0}
								onClick={() => onSelect(artist.artist)}
								onKeyDown={(event) => handleActionAreaKeyDown(event, () => onSelect(artist.artist))}
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
								<CardContent sx={{ p: 1.5, textAlign: 'center', '&:last-child': { pb: 1.5 } }}>
									<Typography variant='subtitle2' fontWeight={600} noWrap>
										{artist.artist}
									</Typography>
									<Typography variant='caption' color='text.secondary'>
										{artist.album_count} albums
									</Typography>
								</CardContent>
								<IconButton
									className='play-overlay'
									onClick={(e) => handlePlayArtist(e, artist.artist)}
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
										'&:hover': { bgcolor: 'primary.light', transform: 'translateY(0) scale(1.05)' },
									}}
								>
									<Play size={16} fill='white' />
								</IconButton>
							</CardActionArea>
						</Card>
					</Grid>
				))}
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
};

const ArtistTracksView = ({ artist, onBack }: { artist: string; onBack: () => void }) => {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const playbackContext = createArtistPlaybackContext(artist);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-artist', artist],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> => {
			return getMusicByArtist(artist, pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const tracks = data?.pages.flatMap((page) => page.items) ?? [];

	const handlePlayAll = () => {
		if (tracks.length > 0) replaceQueue(tracks, 0, playbackContext);
	};

	const handleShuffleAll = () => {
		if (tracks.length > 0) {
			const shuffled = [...tracks].sort(() => Math.random() - 0.5);
			replaceQueue(shuffled, 0, playbackContext);
		}
	};

	return (
		<Box sx={{ p: 2 }}>
			<CategoryHeader
				title={artist}
				trackCount={tracks.length}
				icon={<User size={48} opacity={0.7} />}
				gradientFrom='#4f46e5'
				onBack={onBack}
				onPlayAll={handlePlayAll}
				onShuffleAll={handleShuffleAll}
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
							onPlay={(_, trackIndex) => replaceQueue(tracks, trackIndex, playbackContext)}
							onAddToPlaylist={(e, fileId) => setMenuAnchor({ el: e.currentTarget as HTMLElement, fileId })}
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
};

export default ArtistsView;
