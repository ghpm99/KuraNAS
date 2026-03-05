import { Box, Button, CircularProgress, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { ListMusic, Plus, Trash2 } from 'lucide-react';
import { Playlist } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';

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
				<Typography variant='h6'>{t('MUSIC_PLAYLISTS')}</Typography>
				<Button startIcon={<Plus size={18} />} size='small' variant='contained' onClick={onCreateOpen}>
					{t('MUSIC_NEW')}
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
									onClick={(event) => {
										event.stopPropagation();
										onDelete(playlist.id);
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
								secondary={`${playlist.track_count} ${t('MUSIC_TRACKS_COUNT')}${playlist.description ? ` - ${playlist.description}` : ''}`}
							/>
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
					<Typography variant='body2' sx={{ cursor: 'pointer', color: 'primary.main' }} onClick={onLoadMore}>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}
