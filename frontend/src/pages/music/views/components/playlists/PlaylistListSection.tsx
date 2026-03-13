import { Box, Button, CircularProgress, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { ListMusic, Play, Plus, Trash2 } from 'lucide-react';
import { Playlist } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { getPlaylistTracks } from '@/service/playlist';

type PlaylistListSectionProps = {
	playlists: Playlist[];
	isLoading: boolean;
	hasNextPage: boolean;
	isFetchingNextPage: boolean;
	onSelect: (playlist: Playlist) => void;
	onDelete: (playlistId: number) => void;
	onLoadMore: () => void;
	onCreateOpen: () => void;
};

export default function PlaylistListSection({
	playlists,
	isLoading,
	hasNextPage,
	isFetchingNextPage,
	onSelect,
	onDelete,
	onLoadMore,
	onCreateOpen,
}: PlaylistListSectionProps) {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const handlePlayPlaylist = async (e: React.MouseEvent, playlist: Playlist) => {
		e.stopPropagation();
		const data = await getPlaylistTracks(playlist.id, 1, 200);
		const tracks = data.items.map((item) => item.file);
		if (tracks.length > 0) replaceQueue(tracks);
	};

	if (isLoading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
				<CircularProgress />
			</Box>
		);
	}

	return (
		<Box sx={{ p: 1 }}>
			<Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', p: 1, mb: 1 }}>
				<Typography variant='h6' fontWeight={700}>
					{t('MUSIC_PLAYLISTS')}
				</Typography>
				<Button startIcon={<Plus size={16} />} size='small' variant='contained' onClick={onCreateOpen}>
					{t('MUSIC_NEW')}
				</Button>
			</Box>

			<List sx={{ width: '100%' }}>
				{playlists.map((playlist) => (
					<ListItem
						key={playlist.id}
						disablePadding
						sx={{
							'&:hover .playlist-actions': { opacity: 1 },
						}}
					>
						<ListItemButton
							onClick={() => onSelect(playlist)}
							sx={{ borderRadius: 1.5, py: 1, px: 1.5, gap: 1 }}
						>
							<ListItemIcon sx={{ minWidth: 40 }}>
								<Box
									sx={{
										width: 40,
										height: 40,
										borderRadius: 1,
										bgcolor: playlist.is_system ? 'rgba(167, 139, 250, 0.15)' : 'rgba(99, 102, 241, 0.12)',
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'center',
									}}
								>
									<ListMusic size={20} color={playlist.is_system ? '#a78bfa' : '#6366f1'} />
								</Box>
							</ListItemIcon>
							<ListItemText
								primary={playlist.name}
								secondary={`${playlist.track_count} ${t('MUSIC_TRACKS_COUNT')}${playlist.description ? ` · ${playlist.description}` : ''}`}
								primaryTypographyProps={{ fontWeight: 500 }}
							/>
							<Box className='playlist-actions' sx={{ display: 'flex', gap: 0.5, opacity: 0, transition: 'opacity 0.2s ease' }}>
								<IconButton
									size='small'
									onClick={(e) => handlePlayPlaylist(e, playlist)}
									sx={{ color: 'primary.main' }}
								>
									<Play size={16} fill='#6366f1' />
								</IconButton>
								{!playlist.is_system && (
									<IconButton
										size='small'
										onClick={(e) => {
											e.stopPropagation();
											onDelete(playlist.id);
										}}
										sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
									>
										<Trash2 size={16} />
									</IconButton>
								)}
							</Box>
						</ListItemButton>
					</ListItem>
				))}
			</List>

			{playlists.length === 0 && (
				<Typography variant='body2' color='text.secondary' sx={{ textAlign: 'center', p: 4 }}>
					{t('MUSIC_NO_PLAYLISTS_MSG')}
				</Typography>
			)}

			{hasNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
					<Typography
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main', '&:hover': { textDecoration: 'underline' } }}
						onClick={onLoadMore}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}
