import { Box, CircularProgress, IconButton, List, Typography } from '@mui/material';
import { Play, Shuffle } from 'lucide-react';
import { createAllTracksPlaybackContext } from '@/components/music/playbackContext';
import { useMusic } from '@/components/providers/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import TrackListItem from '@/components/music/TrackListItem';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useState } from 'react';

const AllTracksView = () => {
	const { music, hasNextPage, isFetchingNextPage, lastItemRef } = useMusic();
	const { replaceQueue } = useGlobalMusic();
	const { t } = useI18n();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const playbackContext = createAllTracksPlaybackContext();

	const handlePlayAll = () => {
		if (music.length > 0) replaceQueue(music, 0, playbackContext);
	};

	const handleShuffleAll = () => {
		if (music.length > 0) {
			const shuffled = [...music].sort(() => Math.random() - 0.5);
			replaceQueue(shuffled, 0, playbackContext);
		}
	};

	return (
		<>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, p: 2, pb: 1 }}>
				<Typography variant='h6' fontWeight={700} sx={{ flex: 1 }}>
					{t('MUSIC_ALL_TRACKS')}
				</Typography>
				{music.length > 0 && (
					<Box sx={{ display: 'flex', gap: 1 }}>
						<IconButton
							onClick={handlePlayAll}
							sx={{
								bgcolor: 'primary.main',
								color: 'white',
								width: 36,
								height: 36,
								'&:hover': { bgcolor: 'primary.light', transform: 'scale(1.05)' },
								transition: 'all 0.2s ease',
							}}
						>
							<Play size={18} fill='white' />
						</IconButton>
						<IconButton onClick={handleShuffleAll} sx={{ color: 'text.secondary', '&:hover': { color: 'text.primary' } }}>
							<Shuffle size={18} />
						</IconButton>
					</Box>
				)}
			</Box>

			<List sx={{ width: '100%', px: 1 }}>
				{music.map((item, index) => {
					const isLastItem = index === music.length - 1;
					return (
						<Box key={item.id} ref={isLastItem ? lastItemRef : null}>
							<TrackListItem
								track={item}
								index={index}
								onPlay={(_, trackIndex) => replaceQueue(music, trackIndex, playbackContext)}
								onAddToPlaylist={(e, fileId) => setMenuAnchor({ el: e.currentTarget as HTMLElement, fileId })}
							/>
						</Box>
					);
				})}
			</List>

			<AddToPlaylistMenu
				fileId={menuAnchor?.fileId ?? 0}
				anchorEl={menuAnchor?.el ?? null}
				onClose={() => setMenuAnchor(null)}
			/>

			{isFetchingNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
					<CircularProgress size={32} />
				</Box>
			)}

			{!hasNextPage && music.length > 0 && (
				<Typography variant='caption' color='text.secondary' sx={{ display: 'block', textAlign: 'center', p: 2 }}>
					{t('MUSIC_ALL_LOADED')}
				</Typography>
			)}
		</>
	);
};

export default AllTracksView;
