import { Box, Card, CardActionArea, CardContent, CircularProgress, Grid, IconButton, List, Typography } from '@mui/material';
import { Disc, Play } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { getMusicAlbums, getMusicByAlbum } from '@/service/music';
import { createAlbumPlaybackContext } from '@/components/music/playbackContext';
import { MusicAlbum } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import TrackListItem from '@/components/music/TrackListItem';
import CategoryHeader from '@/components/music/CategoryHeader';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useSearchParams } from 'react-router-dom';

const AlbumsView = () => {
	const [selectedAlbum, setSelectedAlbum] = useState<MusicAlbum | null>(null);
	const [searchParams, setSearchParams] = useSearchParams();
	const selectedAlbumKey = searchParams.get('album') ?? '';

	const handleSelectAlbum = (album: MusicAlbum) => {
		setSelectedAlbum(album);
		setSearchParams((current) => {
			const next = new URLSearchParams(current);
			next.set('album', album.key);
			return next;
		}, { replace: true });
	};

	const handleBack = () => {
		setSelectedAlbum(null);
		setSearchParams((current) => {
			const next = new URLSearchParams(current);
			next.delete('album');
			return next;
		}, { replace: true });
	};

	if (selectedAlbum) {
		return (
			<AlbumTracksView
				album={selectedAlbum}
				onBack={handleBack}
			/>
		);
	}

	return <AlbumListView onSelect={handleSelectAlbum} selectedAlbumKey={selectedAlbumKey} />;
};

const handleActionAreaKeyDown = (event: React.KeyboardEvent<HTMLElement>, onActivate: () => void) => {
	if (event.key === 'Enter' || event.key === ' ') {
		event.preventDefault();
		onActivate();
	}
};

const AlbumListView = ({ onSelect, selectedAlbumKey }: { onSelect: (album: MusicAlbum) => void; selectedAlbumKey: string }) => {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-albums'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicAlbum>> => {
			return getMusicAlbums(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const albums = data?.pages.flatMap((page) => page.items) ?? [];

	useEffect(() => {
		if (!selectedAlbumKey) {
			return;
		}

		const requestedAlbum = albums.find((album) => album.key === selectedAlbumKey);
		if (requestedAlbum) {
			onSelect(requestedAlbum);
		}
	}, [albums, onSelect, selectedAlbumKey]);

	const handlePlayAlbum = async (e: React.MouseEvent, album: MusicAlbum) => {
		e.stopPropagation();
		const data = await getMusicByAlbum(album.key, 1, 200);
		if (data.items.length > 0) replaceQueue(data.items, 0, createAlbumPlaybackContext(album.album));
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
					<Grid key={`${album.album}-${album.artist}`} size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}>
						<Card
							sx={{
								bgcolor: 'background.paper',
								transition: 'all 0.2s ease',
								'&:hover': { bgcolor: 'rgba(255,255,255,0.04)' },
								'&:hover .play-overlay': { opacity: 1, transform: 'translateY(0)' },
							}}
						>
							<CardActionArea
								component='div'
								role='button'
								tabIndex={0}
								onClick={() => onSelect(album)}
								onKeyDown={(event) => handleActionAreaKeyDown(event, () => onSelect(album))}
								sx={{ position: 'relative' }}
							>
								<Box
									sx={{
										height: 140,
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'center',
										bgcolor: 'secondary.dark',
										background: 'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)',
									}}
								>
									<Disc size={48} opacity={0.5} />
								</Box>
								<CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
									<Typography variant='subtitle2' fontWeight={600} noWrap>
										{album.album}
									</Typography>
									<Typography variant='caption' color='text.secondary' noWrap component='div'>
										{album.artist} {album.year ? `· ${album.year}` : ''}
									</Typography>
								</CardContent>
								<IconButton
									className='play-overlay'
									onClick={(e) => handlePlayAlbum(e, album)}
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

const AlbumTracksView = ({ album, onBack }: { album: MusicAlbum; onBack: () => void }) => {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const playbackContext = createAlbumPlaybackContext(album.album);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-album', album.key],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> => {
			return getMusicByAlbum(album.key, pageParam, 50);
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
				title={album.album}
				subtitle={album.artist}
				trackCount={tracks.length}
				icon={<Disc size={48} opacity={0.7} />}
				gradientFrom='#7c3aed'
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

export default AlbumsView;
