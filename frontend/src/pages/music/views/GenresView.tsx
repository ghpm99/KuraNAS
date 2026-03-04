import {
	Box,
	CircularProgress,
	IconButton,
	List,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	Typography,
} from '@mui/material';
import { ArrowLeft, ListPlus, Music, Play, Tag } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { getMusicGenres, getMusicByGenre } from '@/service/music';
import { MusicGenre } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';

const GenresView = () => {
	const [selectedGenre, setSelectedGenre] = useState<string | null>(null);

	if (selectedGenre) {
		return <GenreTracksView genre={selectedGenre} onBack={() => setSelectedGenre(null)} />;
	}

	return <GenreListView onSelect={setSelectedGenre} />;
};

const GenreListView = ({ onSelect }: { onSelect: (genre: string) => void }) => {
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-genres'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicGenre>> => {
			return getMusicGenres(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const genres = data?.pages.flatMap((page) => page.items) ?? [];

	if (isLoading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
				<CircularProgress />
			</Box>
		);
	}

	return (
		<List sx={{ width: '100%' }}>
			{genres.map((genre) => (
				<ListItem key={genre.genre} sx={{ px: 0 }}>
					<ListItemButton onClick={() => onSelect(genre.genre)}>
						<ListItemIcon>
							<Tag />
						</ListItemIcon>
						<ListItemText primary={genre.genre} secondary={`${genre.track_count} tracks`} />
					</ListItemButton>
				</ListItem>
			))}

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
		</List>
	);
};

const GenreTracksView = ({ genre, onBack }: { genre: string; onBack: () => void }) => {
	const { getMusicTitle, musicMetadata, getMusicArtist, addToQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-genre', genre],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> => {
			return getMusicByGenre(genre, pageParam, 50);
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
				<Typography variant='h6'>{genre}</Typography>
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

export default GenresView;
