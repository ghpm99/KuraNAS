import { Box, CircularProgress, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { Folder, Play } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { getMusicByFolder, getMusicFolders } from '@/service/music';
import { createFolderPlaybackContext } from '@/components/music/playbackContext';
import { MusicFolder } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import TrackListItem from '@/components/music/TrackListItem';
import CategoryHeader from '@/components/music/CategoryHeader';
import useI18n from '@/components/i18n/provider/i18nContext';

const FoldersView = () => {
	const [selectedFolder, setSelectedFolder] = useState<string | null>(null);

	if (selectedFolder) {
		return <FolderTracksView folder={selectedFolder} onBack={() => setSelectedFolder(null)} />;
	}

	return <FolderListView onSelect={setSelectedFolder} />;
};

const handleListItemKeyDown = (event: React.KeyboardEvent<HTMLElement>, onActivate: () => void) => {
	if (event.key === 'Enter' || event.key === ' ') {
		event.preventDefault();
		onActivate();
	}
};

const FolderListView = ({ onSelect }: { onSelect: (folder: string) => void }) => {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-folders'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicFolder>> => {
			return getMusicFolders(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const folders = data?.pages.flatMap((page) => page.items) ?? [];

	const getFolderName = (path: string) => {
		const parts = path.split('/').filter(Boolean);
		return parts[parts.length - 1] || path;
	};

	const handlePlayFolder = async (e: React.MouseEvent, folder: string) => {
		e.stopPropagation();
		const data = await getMusicByFolder(folder, 1, 500);
		const tracks = data.items;
		if (tracks.length > 0) replaceQueue(tracks, 0, createFolderPlaybackContext(folder));
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
			<List sx={{ width: '100%' }}>
				{folders.map((folder) => (
					<ListItem
						key={folder.folder}
						disablePadding
						sx={{
							'&:hover .folder-play': { opacity: 1 },
						}}
					>
						<ListItemButton
							component='div'
							role='button'
							tabIndex={0}
							onClick={() => onSelect(folder.folder)}
							onKeyDown={(event) => handleListItemKeyDown(event, () => onSelect(folder.folder))}
							sx={{ borderRadius: 1.5, py: 1, px: 1.5, gap: 1 }}
						>
							<ListItemIcon sx={{ minWidth: 40 }}>
								<Box
									sx={{
										width: 40,
										height: 40,
										borderRadius: 1,
										bgcolor: 'rgba(99, 102, 241, 0.12)',
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'center',
									}}
								>
									<Folder size={20} color='#6366f1' />
								</Box>
							</ListItemIcon>
							<ListItemText
								primary={getFolderName(folder.folder)}
								secondary={`${folder.track_count} ${t('MUSIC_TRACKS_COUNT')}`}
								primaryTypographyProps={{ fontWeight: 500 }}
							/>
							<IconButton
								className='folder-play'
								onClick={(e) => handlePlayFolder(e, folder.folder)}
								sx={{
									opacity: 0,
									transition: 'all 0.2s ease',
									color: 'primary.main',
									'&:hover': { bgcolor: 'rgba(99, 102, 241, 0.12)' },
								}}
							>
								<Play size={18} fill='#6366f1' />
							</IconButton>
						</ListItemButton>
					</ListItem>
				))}
			</List>

			{hasNextPage && (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
					<Typography
						variant='body2'
						sx={{ cursor: 'pointer', color: 'primary.main', '&:hover': { textDecoration: 'underline' } }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
};

const FolderTracksView = ({ folder, onBack }: { folder: string; onBack: () => void }) => {
	const { replaceQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const playbackContext = createFolderPlaybackContext(folder);

	const { data, isLoading } = useQuery({
		queryKey: ['music-by-folder', folder],
		queryFn: (): Promise<Pagination<IMusicData>> => getMusicByFolder(folder, 1, 500),
	});

	const tracks = data?.items ?? [];

	const getFolderName = (path: string) => {
		const parts = path.split('/').filter(Boolean);
		return parts[parts.length - 1] || path;
	};

	const handlePlayAll = () => {
		if (tracks.length > 0) replaceQueue(tracks, 0, playbackContext);
	};

	const handleShuffleAll = () => {
		if (tracks.length > 0) {
			const shuffled = [...tracks].sort(() => Math.random() - 0.5);
			replaceQueue(shuffled, 0, playbackContext);
		}
	};

	return (
		<Box sx={{ p: 2 }}>
			<CategoryHeader
				title={getFolderName(folder)}
				subtitle={folder}
				trackCount={tracks.length}
				icon={<Folder size={48} opacity={0.7} />}
				gradientFrom='#6366f1'
				onBack={onBack}
				onPlayAll={handlePlayAll}
				onShuffleAll={handleShuffleAll}
			/>

			{isLoading ? (
				<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
					<CircularProgress />
				</Box>
			) : (
				<List sx={{ width: '100%' }}>
					{tracks.map((item, index) => (
						<TrackListItem
							key={item.id}
							track={item}
							index={index}
							onPlay={(_, trackIndex) => replaceQueue(tracks, trackIndex, playbackContext)}
							onAddToPlaylist={(e, fileId) => setMenuAnchor({ el: e.currentTarget as HTMLElement, fileId })}
						/>
					))}
				</List>
			)}

			<AddToPlaylistMenu
				fileId={menuAnchor?.fileId ?? 0}
				anchorEl={menuAnchor?.el ?? null}
				onClose={() => setMenuAnchor(null)}
			/>
		</Box>
	);
};

export default FoldersView;
