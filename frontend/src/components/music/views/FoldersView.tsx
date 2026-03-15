import { Box, CircularProgress, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { Folder, Play } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import CategoryHeader from '@/components/music/CategoryHeader';
import TrackListItem from '@/components/music/TrackListItem';
import { createFolderPlaybackContext } from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicByFolder, getMusicFolders } from '@/service/music';
import { MusicFolder } from '@/types/music';
import { Pagination } from '@/types/pagination';
import {
	getFolderName,
	handleKeyboardActivation,
	loadAllTracks,
	MUSIC_COLLECTION_PAGE_SIZE,
	shuffleTracks,
} from './shared';

const loadFolderTracks = (folderPath: string) =>
	loadAllTracks((page, pageSize) => getMusicByFolder(folderPath, page, pageSize));

export default function FoldersView() {
	const [searchParams, setSearchParams] = useSearchParams();
	const selectedFolderPath = searchParams.get('folder') ?? '';
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-folders'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicFolder>> => getMusicFolders(pageParam, MUSIC_COLLECTION_PAGE_SIZE),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});
	const folders = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data]);

	const handleSelectFolder = (folder: string) => {
		setSearchParams((current) => {
			const next = new URLSearchParams(current);
			next.set('folder', folder);
			return next;
		}, { replace: true });
	};

	const handleBack = () => {
		setSearchParams((current) => {
			const next = new URLSearchParams(current);
			next.delete('folder');
			return next;
		}, { replace: true });
	};

	if (selectedFolderPath) {
		return <FolderTracksView folder={selectedFolderPath} onBack={handleBack} />;
	}

	return (
		<FolderListView
			folders={folders}
			isLoading={isLoading}
			fetchNextPage={fetchNextPage}
			hasNextPage={hasNextPage}
			isFetchingNextPage={isFetchingNextPage}
			onSelect={handleSelectFolder}
		/>
	);
}

type FolderListViewProps = {
	folders: MusicFolder[];
	isLoading: boolean;
	fetchNextPage: () => Promise<unknown>;
	hasNextPage: boolean;
	isFetchingNextPage: boolean;
	onSelect: (folder: string) => void;
};

function FolderListView({ folders, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage, onSelect }: FolderListViewProps) {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();

	const handlePlayFolder = async (event: React.MouseEvent, folderPath: string) => {
		event.stopPropagation();
		const tracks = await loadFolderTracks(folderPath);
		if (tracks.length > 0) {
			replaceQueue(tracks, 0, createFolderPlaybackContext(folderPath));
		}
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
							onKeyDown={(event) => handleKeyboardActivation(event, () => onSelect(folder.folder))}
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
								onClick={(event) => void handlePlayFolder(event, folder.folder)}
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
}

function FolderTracksView({ folder, onBack }: { folder: string; onBack: () => void }) {
	const { t } = useI18n();
	const { replaceQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);
	const playbackContext = createFolderPlaybackContext(folder);

	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-by-folder', folder],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<IMusicData>> =>
			getMusicByFolder(folder, pageParam, MUSIC_COLLECTION_PAGE_SIZE),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const tracks = data?.pages.flatMap((page) => page.items) ?? [];

	const queueFolderTracks = async (trackId?: number, shuffle = false) => {
		const allTracks = await loadFolderTracks(folder);
		if (allTracks.length === 0) {
			return;
		}

		if (shuffle) {
			replaceQueue(shuffleTracks(allTracks), 0, playbackContext);
			return;
		}

		const startIndex = trackId ? Math.max(allTracks.findIndex((item) => item.id === trackId), 0) : 0;
		replaceQueue(allTracks, startIndex, playbackContext);
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
				onPlayAll={() => void queueFolderTracks()}
				onShuffleAll={() => void queueFolderTracks(undefined, true)}
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
							onPlay={(track) => void queueFolderTracks(track.id)}
							onAddToPlaylist={(event, fileId) => setMenuAnchor({ el: event.currentTarget as HTMLElement, fileId })}
						/>
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
						sx={{ cursor: 'pointer', color: 'primary.main', '&:hover': { textDecoration: 'underline' } }}
						onClick={() => fetchNextPage()}
					>
						{isFetchingNextPage ? <CircularProgress size={20} /> : t('ACTION_LOAD_MORE')}
					</Typography>
				</Box>
			)}
		</Box>
	);
}
