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
import { ArrowLeft, ListPlus, Music, Play, User } from 'lucide-react';
import { useState } from 'react';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import { useInfiniteQuery } from '@tanstack/react-query';
import { getMusicArtists, getMusicByArtist } from '@/service/music';
import { MusicArtist } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

const ArtistsView = () => {
	const [selectedArtist, setSelectedArtist] = useState<string | null>(null);

	if (selectedArtist) {
		return <ArtistTracksView artist={selectedArtist} onBack={() => setSelectedArtist(null)} />;
	}

	return <ArtistListView onSelect={setSelectedArtist} />;
};

const ArtistListView = ({ onSelect }: { onSelect: (artist: string) => void }) => {
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-artists'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicArtist>> => {
			return getMusicArtists(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const artists = data?.pages.flatMap((page) => page.items) ?? [];

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
				{artists.map((artist) => (
					<Grid key={artist.artist} size={{ xs: 6, sm: 4, md: 3 }}>
						<Card sx={{ bgcolor: 'background.paper' }}>
							<CardActionArea onClick={() => onSelect(artist.artist)}>
								<Box
									sx={{
										height: 120,
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'center',
										bgcolor: 'primary.dark',
									}}
								>
									<User size={48} opacity={0.6} />
								</Box>
								<CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
									<Typography variant='subtitle2' noWrap>
										{artist.artist}
									</Typography>
									<Typography variant='caption' color='text.secondary'>
										{artist.album_count} albums - {artist.track_count} tracks
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

const ArtistTracksView = ({ artist, onBack }: { artist: string; onBack: () => void }) => {
	const { getMusicTitle, musicMetadata, addToQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-artist', artist],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> => {
			return getMusicByArtist(artist, pageParam, 50);
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
				<Typography variant='h6'>{artist}</Typography>
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
								<ListItemText primary={getMusicTitle(item)} secondary={musicMetadata(item)} />
								<IconButton
									sx={{ color: 'rgba(255, 255, 255, 0.4)' }}
									aria-label={`add ${item.name} to playlist`}
									onClick={(e) => {
										e.stopPropagation();
										setMenuAnchor({ el: e.currentTarget, fileId: item.id });
									}}
								>
									<ListPlus size={18} />
								</IconButton>
								<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }}>
									<Play />
								</IconButton>
							</ListItemButton>
						</ListItem>
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

export default ArtistsView;
