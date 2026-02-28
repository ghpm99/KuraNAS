import {
	Box,
	Button,
	CircularProgress,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	IconButton,
	List,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	TextField,
	Typography,
} from '@mui/material';
import { ArrowLeft, ListMusic, Music, Play, Plus, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import {
	getPlaylists,
	createPlaylist,
	deletePlaylist,
	getPlaylistTracks,
	removeTrackFromPlaylist,
} from '@/service/playlist';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { useSnackbar } from 'notistack';

const PlaylistsView = () => {
	const [selectedPlaylist, setSelectedPlaylist] = useState<Playlist | null>(null);

	if (selectedPlaylist) {
		return <PlaylistDetailView playlist={selectedPlaylist} onBack={() => setSelectedPlaylist(null)} />;
	}

	return <PlaylistListView onSelect={setSelectedPlaylist} />;
};

const PlaylistListView = ({ onSelect }: { onSelect: (playlist: Playlist) => void }) => {
	const [createOpen, setCreateOpen] = useState(false);
	const [newName, setNewName] = useState('');
	const [newDescription, setNewDescription] = useState('');
	const queryClient = useQueryClient();
	const { enqueueSnackbar } = useSnackbar();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['playlists'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<Playlist>> => {
			return getPlaylists(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const createMutation = useMutation({
		mutationFn: () => createPlaylist({ name: newName, description: newDescription }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			setCreateOpen(false);
			setNewName('');
			setNewDescription('');
			enqueueSnackbar('Playlist created', { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar('Failed to create playlist', { variant: 'error' });
		},
	});

	const deleteMutation = useMutation({
		mutationFn: (id: number) => deletePlaylist(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			enqueueSnackbar('Playlist deleted', { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar('Failed to delete playlist', { variant: 'error' });
		},
	});

	const playlists = data?.pages.flatMap((page) => page.items) ?? [];

	if (isLoading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
				<CircularProgress />
			</Box>
		);
	}

	return (
		<Box>
			<Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', p: 1 }}>
				<Typography variant='h6'>Playlists</Typography>
				<Button
					startIcon={<Plus size={18} />}
					size='small'
					variant='contained'
					onClick={() => setCreateOpen(true)}
				>
					New
				</Button>
			</Box>

			<List sx={{ width: '100%' }}>
				{playlists.map((playlist) => (
					<ListItem
						key={playlist.id}
						sx={{ px: 0 }}
						secondaryAction={
							!playlist.is_system ? (
								<IconButton
									edge='end'
									onClick={(e) => {
										e.stopPropagation();
										deleteMutation.mutate(playlist.id);
									}}
									sx={{ color: 'rgba(255, 255, 255, 0.4)', '&:hover': { color: 'error.main' } }}
								>
									<Trash2 size={18} />
								</IconButton>
							) : undefined
						}
					>
						<ListItemButton onClick={() => onSelect(playlist)}>
							<ListItemIcon>
								<ListMusic />
							</ListItemIcon>
							<ListItemText
								primary={playlist.name}
								secondary={`${playlist.track_count} tracks${playlist.description ? ` - ${playlist.description}` : ''}`}
							/>
						</ListItemButton>
					</ListItem>
				))}
			</List>

			{playlists.length === 0 && (
				<Typography variant='body2' color='text.secondary' sx={{ textAlign: 'center', p: 4 }}>
					No playlists yet. Create one to get started.
				</Typography>
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

			<Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth='sm' fullWidth>
				<DialogTitle>Create Playlist</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						fullWidth
						label='Name'
						value={newName}
						onChange={(e) => setNewName(e.target.value)}
						sx={{ mt: 1, mb: 2 }}
					/>
					<TextField
						fullWidth
						label='Description (optional)'
						value={newDescription}
						onChange={(e) => setNewDescription(e.target.value)}
						multiline
						rows={2}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setCreateOpen(false)}>Cancel</Button>
					<Button
						variant='contained'
						onClick={() => createMutation.mutate()}
						disabled={!newName.trim() || createMutation.isPending}
					>
						{createMutation.isPending ? <CircularProgress size={20} /> : 'Create'}
					</Button>
				</DialogActions>
			</Dialog>
		</Box>
	);
};

const PlaylistDetailView = ({ playlist, onBack }: { playlist: Playlist; onBack: () => void }) => {
	const { getMusicTitle, getMusicArtist, musicMetadata, addToQueue } = useGlobalMusic();
	const queryClient = useQueryClient();
	const { enqueueSnackbar } = useSnackbar();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['playlist-tracks', playlist.id],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<PlaylistTrack>> => {
			return getPlaylistTracks(playlist.id, pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const removeMutation = useMutation({
		mutationFn: (fileId: number) => removeTrackFromPlaylist(playlist.id, fileId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlist-tracks', playlist.id] });
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			enqueueSnackbar('Track removed', { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar('Failed to remove track', { variant: 'error' });
		},
	});

	const tracks = data?.pages.flatMap((page) => page.items) ?? [];

	return (
		<Box>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 1 }}>
				<IconButton onClick={onBack} size='small'>
					<ArrowLeft />
				</IconButton>
				<Box>
					<Typography variant='h6'>{playlist.name}</Typography>
					{playlist.description && (
						<Typography variant='caption' color='text.secondary'>
							{playlist.description}
						</Typography>
					)}
				</Box>
				<Typography variant='caption' color='text.secondary' sx={{ ml: 1 }}>
					({tracks.length} tracks)
				</Typography>
			</Box>

			{isLoading ? (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
					<CircularProgress />
				</Box>
			) : tracks.length === 0 ? (
				<Typography variant='body2' color='text.secondary' sx={{ textAlign: 'center', p: 4 }}>
					This playlist is empty. Add tracks from any view.
				</Typography>
			) : (
				<List sx={{ width: '100%' }}>
					{tracks.map((track) => (
						<ListItem
							key={track.id}
							sx={{ px: 0 }}
							secondaryAction={
								<IconButton
									edge='end'
									onClick={() => removeMutation.mutate(track.file.id)}
									sx={{ color: 'rgba(255, 255, 255, 0.4)', '&:hover': { color: 'error.main' } }}
								>
									<Trash2 size={16} />
								</IconButton>
							}
						>
							<ListItemButton onClick={() => addToQueue(track.file)}>
								<ListItemIcon>
									<Music />
								</ListItemIcon>
								<ListItemText
									primary={getMusicTitle(track.file)}
									secondary={`${getMusicArtist(track.file)} - ${musicMetadata(track.file)}`}
								/>
								<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)', mr: 1 }}>
									<Play size={18} />
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

export default PlaylistsView;
