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
import { ArrowLeft, Folder, ListPlus, Music, Play } from 'lucide-react';
import { useState } from 'react';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { getMusicFolders } from '@/service/music';
import { MusicFolder } from '@/types/music';
import { Pagination } from '@/types/pagination';
import { IMusicData } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { apiBase } from '@/service';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';

const FoldersView = () => {
	const [selectedFolder, setSelectedFolder] = useState<string | null>(null);

	if (selectedFolder) {
		return <FolderTracksView folder={selectedFolder} onBack={() => setSelectedFolder(null)} />;
	}

	return <FolderListView onSelect={setSelectedFolder} />;
};

const FolderListView = ({ onSelect }: { onSelect: (folder: string) => void }) => {
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music-folders'],
		queryFn: async ({ pageParam = 1 }): Promise<Pagination<MusicFolder>> => {
			return getMusicFolders(pageParam, 50);
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const folders = data?.pages.flatMap((page) => page.items) ?? [];

	if (isLoading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
				<CircularProgress />
			</Box>
		);
	}

	const getFolderName = (path: string) => {
		const parts = path.split('/').filter(Boolean);
		return parts[parts.length - 1] || path;
	};

	return (
		<List sx={{ width: '100%' }}>
			{folders.map((folder) => (
				<ListItem key={folder.folder} sx={{ px: 0 }}>
					<ListItemButton onClick={() => onSelect(folder.folder)}>
						<ListItemIcon>
							<Folder />
						</ListItemIcon>
						<ListItemText
							primary={getFolderName(folder.folder)}
							secondary={`${folder.folder} - ${folder.track_count} tracks`}
						/>
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

const FolderTracksView = ({ folder, onBack }: { folder: string; onBack: () => void }) => {
	const { getMusicTitle, musicMetadata, getMusicArtist, addToQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);

	const { data, isLoading } = useQuery({
		queryKey: ['music-by-folder', folder],
		queryFn: async (): Promise<Pagination<IMusicData>> => {
			const response = await apiBase.get<Pagination<IMusicData>>('/files/music', {
				params: { page: 1, page_size: 500 },
			});
			return {
				...response.data,
				items: response.data.items.filter((item) => item.path.startsWith(folder)),
			};
		},
	});

	const tracks = data?.items ?? [];

	const getFolderName = (path: string) => {
		const parts = path.split('/').filter(Boolean);
		return parts[parts.length - 1] || path;
	};

	return (
		<Box>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 1 }}>
				<IconButton onClick={onBack} size='small'>
					<ArrowLeft />
				</IconButton>
				<Typography variant='h6'>{getFolderName(folder)}</Typography>
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
		</Box>
	);
};

export default FoldersView;
