import { Box, CircularProgress, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { ArrowLeft, Music, Play, Trash2 } from 'lucide-react';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';
import { usePlaylistTrackHandlers } from '@/components/hooks/usePlaylistTrackHandlers/usePlaylistTrackHandlers';

type PlaylistDetailSectionProps = {
	playlist: Playlist;
	tracks: PlaylistTrack[];
	isLoading: boolean;
	hasNextPage: boolean;
	isFetchingNextPage: boolean;
	onBack: () => void;
	onRemoveTrack: (fileId: number) => void;
	onLoadMore: () => void;
};

export default function PlaylistDetailSection({
	playlist,
	tracks,
	isLoading,
	hasNextPage,
	isFetchingNextPage,
	onBack,
	onRemoveTrack,
	onLoadMore,
}: PlaylistDetailSectionProps) {
	const { t } = useI18n();
	const { addToQueue, getMusicArtist, getMusicTitle, musicMetadata } = usePlaylistTrackHandlers();

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
					({tracks.length} {t('MUSIC_TRACKS_COUNT')})
				</Typography>
			</Box>

			{isLoading ? (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
					<CircularProgress />
				</Box>
			) : tracks.length === 0 ? (
				<Typography variant='body2' color='text.secondary' sx={{ textAlign: 'center', p: 4 }}>
					{t('MUSIC_PLAYLIST_EMPTY')}
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
									onClick={() => onRemoveTrack(track.file.id)}
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
					<Typography variant='body2' sx={{ cursor: 'pointer', color: 'primary.main' }} onClick={onLoadMore}>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}
