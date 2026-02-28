import {
	Box,
	Card,
	CardActionArea,
	CardContent,
	CircularProgress,
	Grid,
	IconButton,
	List,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	Typography,
} from '@mui/material';
import { ArrowLeft, Disc, Music, Play } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { getMusicAlbums, getMusicByAlbum } from '@/service/music';
import { MusicAlbum } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

const AlbumsView = () => {
	const [selectedAlbum, setSelectedAlbum] = useState<string | null>(null);

	if (selectedAlbum) {
		return <AlbumTracksView album={selectedAlbum} onBack={() => setSelectedAlbum(null)} />;
	}

	return <AlbumListView onSelect={setSelectedAlbum} />;
};

const AlbumListView = ({ onSelect }: { onSelect: (album: string) => void }) => {
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-albums'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicAlbum>> => {
			return getMusicAlbums(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const albums = data?.pages.flatMap((page) => page.items) ?? [];

	if (isLoading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
				<CircularProgress />
			</Box>
		);
	}

	return (
		<Box sx={{ p: 1 }}>
			<Grid container spacing={2}>
				{albums.map((album) => (
					<Grid key={`${album.album}-${album.artist}`} size={{ xs: 6, sm: 4, md: 3 }}>
						<Card sx={{ bgcolor: 'background.paper' }}>
							<CardActionArea onClick={() => onSelect(album.album)}>
								<Box
									sx={{
										height: 120,
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'center',
										bgcolor: 'secondary.dark',
									}}
								>
									<Disc size={48} opacity={0.6} />
								</Box>
								<CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
									<Typography variant='subtitle2' noWrap>
										{album.album}
									</Typography>
									<Typography variant='caption' color='text.secondary' noWrap>
										{album.artist} {album.year ? `(${album.year})` : ''} - {album.track_count} tracks
									</Typography>
								</CardContent>
							</CardActionArea>
						</Card>
					</Grid>
				))}
			</Grid>

			{hasNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
					<Typography
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main' }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : 'Load more'}
					</Typography>
				</Box>
			)}
		</Box>
	);
};

const AlbumTracksView = ({ album, onBack }: { album: string; onBack: () => void }) => {
	const { getMusicTitle, musicMetadata, getMusicArtist, addToQueue } = useGlobalMusic();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-album', album],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> => {
			return getMusicByAlbum(album, pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const tracks = data?.pages.flatMap((page) => page.items) ?? [];

	return (
		<Box>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 1 }}>
				<IconButton onClick={onBack} size='small'>
					<ArrowLeft />
				</IconButton>
				<Typography variant='h6'>{album}</Typography>
				<Typography variant='caption' color='text.secondary'>
					({tracks.length} tracks)
				</Typography>
			</Box>

			{isLoading ? (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
					<CircularProgress />
				</Box>
			) : (
				<List sx={{ width: '100%' }}>
					{tracks.map((item) => (
						<ListItem key={item.id} sx={{ px: 0 }}>
							<ListItemButton onClick={() => addToQueue(item)}>
								<ListItemIcon>
									<Music />
								</ListItemIcon>
								<ListItemText
									primary={getMusicTitle(item)}
									secondary={`${getMusicArtist(item)} - ${musicMetadata(item)}`}
								/>
								<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }}>
									<Play />
								</IconButton>
							</ListItemButton>
						</ListItem>
					))}
				</List>
			)}

			{hasNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
					<Typography
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main' }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : 'Load more'}
					</Typography>
				</Box>
			)}
		</Box>
	);
};

export default AlbumsView;
