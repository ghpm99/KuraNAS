import {
	Button,
	CircularProgress,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	Menu,
	MenuItem,
	TextField,
} from '@mui/material';
import { ListPlus, ListMusic, Plus } from 'lucide-react';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getPlaylists, addTrackToPlaylist, createPlaylist } from '@/service/playlist';
import { useSnackbar } from 'notistack';

interface AddToPlaylistMenuProps {
	fileId: number;
	anchorEl: HTMLElement | null;
	onClose: () => void;
}

const AddToPlaylistMenu = ({ fileId, anchorEl, onClose }: AddToPlaylistMenuProps) => {
	const [createOpen, setCreateOpen] = useState(false);
	const [newName, setNewName] = useState('');
	const queryClient = useQueryClient();
	const { enqueueSnackbar } = useSnackbar();

	const { data, isLoading } = useQuery({
		queryKey: ['playlists-menu'],
		queryFn: () => getPlaylists(1, 100),
		enabled: !!anchorEl,
	});

	const addMutation = useMutation({
		mutationFn: (playlistId: number) => addTrackToPlaylist(playlistId, fileId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			queryClient.invalidateQueries({ queryKey: ['playlist-tracks'] });
			enqueueSnackbar('Track added to playlist', { variant: 'success' });
			onClose();
		},
		onError: () => {
			enqueueSnackbar('Failed to add track (may already exist)', { variant: 'warning' });
		},
	});

	const createAndAddMutation = useMutation({
		mutationFn: async () => {
			const playlist = await createPlaylist({ name: newName });
			await addTrackToPlaylist(playlist.id, fileId);
			return playlist;
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			queryClient.invalidateQueries({ queryKey: ['playlists-menu'] });
			setCreateOpen(false);
			setNewName('');
			enqueueSnackbar('Playlist created and track added', { variant: 'success' });
			onClose();
		},
		onError: () => {
			enqueueSnackbar('Failed to create playlist', { variant: 'error' });
		},
	});

	const playlists = data?.items?.filter((p) => !p.is_system) ?? [];

	return (
		<>
			<Menu anchorEl={anchorEl} open={!!anchorEl} onClose={onClose}>
				{isLoading ? (
					<MenuItem disabled>
						<CircularProgress size={20} sx={{ mr: 1 }} /> Loading...
					</MenuItem>
				) : (
					<>
						<MenuItem
							onClick={() => {
								setCreateOpen(true);
								onClose();
							}}
						>
							<ListItemIcon>
								<Plus size={18} />
							</ListItemIcon>
							<ListItemText primary='New playlist...' />
						</MenuItem>
						{playlists.map((playlist) => (
							<MenuItem
								key={playlist.id}
								onClick={() => addMutation.mutate(playlist.id)}
								disabled={addMutation.isPending}
							>
								<ListItemIcon>
									<ListMusic size={18} />
								</ListItemIcon>
								<ListItemText primary={playlist.name} />
							</MenuItem>
						))}
						{playlists.length === 0 && (
							<MenuItem disabled>
								<ListItemText primary='No playlists yet' />
							</MenuItem>
						)}
					</>
				)}
			</Menu>

			<Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth='sm' fullWidth>
				<DialogTitle>Create Playlist & Add Track</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						fullWidth
						label='Playlist Name'
						value={newName}
						onChange={(e) => setNewName(e.target.value)}
						sx={{ mt: 1 }}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setCreateOpen(false)}>Cancel</Button>
					<Button
						variant='contained'
						onClick={() => createAndAddMutation.mutate()}
						disabled={!newName.trim() || createAndAddMutation.isPending}
					>
						{createAndAddMutation.isPending ? <CircularProgress size={20} /> : 'Create & Add'}
					</Button>
				</DialogActions>
			</Dialog>
		</>
	);
};

export default AddToPlaylistMenu;

export const AddToPlaylistButton = ({ fileId }: { fileId: number }) => {
	const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

	return (
		<>
			<ListItemButton
				sx={{ px: 1, py: 0.5, borderRadius: 1, maxWidth: 'fit-content' }}
				onClick={(e) => {
					e.stopPropagation();
					setAnchorEl(e.currentTarget);
				}}
			>
				<ListItemIcon sx={{ minWidth: 28 }}>
					<ListPlus size={16} />
				</ListItemIcon>
			</ListItemButton>
			<AddToPlaylistMenu fileId={fileId} anchorEl={anchorEl} onClose={() => setAnchorEl(null)} />
		</>
	);
};
