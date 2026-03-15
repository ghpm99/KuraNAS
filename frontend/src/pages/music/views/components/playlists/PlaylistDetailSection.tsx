import { Box, CircularProgress, IconButton, List, ListItem, ListItemButton, Typography } from '@mui/material';
import { ListMusic, Pause, Play, Trash2 } from 'lucide-react';
import { createPlaylistPlaybackContext } from '@/components/music/playbackContext';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';
import { usePlaylistTrackHandlers } from '@/components/hooks/usePlaylistTrackHandlers/usePlaylistTrackHandlers';
import CategoryHeader from '@/components/music/CategoryHeader';
import { formatMusicDuration } from '@/utils/music';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

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
	const { getMusicArtist, getMusicTitle } = usePlaylistTrackHandlers();
	const { currentTrack, isPlaying, replaceQueue } = useGlobalMusic();
	const handleListItemKeyDown = (event: React.KeyboardEvent<HTMLElement>, onActivate: () => void) => {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			onActivate();
		}
	};

	const allFiles = tracks.map((track) => track.file);
	const playbackContext = createPlaylistPlaybackContext(playlist);
	const canRemoveTracks = !playlist.is_system && !playlist.is_auto;

	const handlePlayAll = () => {
		if (allFiles.length > 0) replaceQueue(allFiles, 0, playbackContext);
	};

	const handleShuffleAll = () => {
		if (allFiles.length > 0) {
			const shuffled = [...allFiles].sort(() => Math.random() - 0.5);
			replaceQueue(shuffled, 0, playbackContext);
		}
	};

	return (
		<Box sx={{ p: 2 }}>
			<CategoryHeader
				title={playlist.name}
				subtitle={playlist.description || undefined}
				trackCount={tracks.length}
				icon={<ListMusic size={48} opacity={0.7} />}
				gradientFrom='#6366f1'
				onBack={onBack}
				onPlayAll={handlePlayAll}
				onShuffleAll={handleShuffleAll}
			/>

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
					{tracks.map((track, index) => {
						const isCurrentTrack = currentTrack?.id === track.file.id;
						const duration = track.file.metadata?.duration;

						return (
							<ListItem
								key={track.id}
								disablePadding
								sx={{
									'&:hover .remove-btn': { opacity: 1 },
								}}
							>
								<ListItemButton
									component='div'
									role='button'
									tabIndex={0}
									onClick={() => replaceQueue(allFiles, index, playbackContext)}
									onKeyDown={(event) => handleListItemKeyDown(event, () => replaceQueue(allFiles, index, playbackContext))}
									sx={{
										borderRadius: 1,
										py: 0.5,
										px: 1,
										gap: 1.5,
										bgcolor: isCurrentTrack ? 'rgba(99, 102, 241, 0.08)' : 'transparent',
										'&:hover': {
											bgcolor: isCurrentTrack ? 'rgba(99, 102, 241, 0.12)' : undefined,
										},
										'&:hover .track-index': { display: 'none' },
										'&:hover .track-play-icon': { display: 'flex' },
									}}
								>
									<Box sx={{ width: 32, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
										{isCurrentTrack && isPlaying ? (
											<Box sx={{ display: 'flex', alignItems: 'flex-end', gap: '2px', height: 16 }}>
												{[1, 2, 3].map((bar) => (
													<Box
														key={bar}
														sx={{
															width: 3,
															bgcolor: 'primary.main',
															borderRadius: 1,
															animation: `equalizer ${0.4 + bar * 0.15}s ease-in-out infinite alternate`,
															'@keyframes equalizer': {
																'0%': { height: '4px' },
																'100%': { height: '14px' },
															},
														}}
													/>
												))}
											</Box>
										) : isCurrentTrack ? (
											<Pause size={14} color='#6366f1' />
										) : (
											<>
												<Typography className='track-index' variant='body2' color='text.secondary' sx={{ fontVariantNumeric: 'tabular-nums' }}>
													{index + 1}
												</Typography>
												<Box className='track-play-icon' sx={{ display: 'none', alignItems: 'center' }}>
													<Play size={14} />
												</Box>
											</>
										)}
									</Box>

									<Box sx={{ flex: 1, minWidth: 0 }}>
										<Typography
											variant='body2'
											noWrap
											fontWeight={isCurrentTrack ? 600 : 400}
											color={isCurrentTrack ? 'primary.main' : 'text.primary'}
										>
											{getMusicTitle(track.file)}
										</Typography>
										<Typography variant='caption' color='text.secondary' noWrap component='div'>
											{getMusicArtist(track.file)}
										</Typography>
									</Box>

									{canRemoveTracks && (
										<IconButton
											className='remove-btn'
											size='small'
											onClick={(e) => {
												e.stopPropagation();
												onRemoveTrack(track.file.id);
											}}
											sx={{ opacity: 0, color: 'text.secondary', '&:hover': { color: 'error.main' } }}
										>
											<Trash2 size={14} />
										</IconButton>
									)}

									{duration ? (
										<Typography variant='caption' color='text.secondary' sx={{ flexShrink: 0, fontVariantNumeric: 'tabular-nums', minWidth: 36, textAlign: 'right' }}>
											{formatMusicDuration(duration)}
										</Typography>
									) : (
										<Box sx={{ width: 36 }} />
									)}
								</ListItemButton>
							</ListItem>
						);
					})}
				</List>
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
