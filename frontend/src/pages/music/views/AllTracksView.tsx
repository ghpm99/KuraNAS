import {
	CircularProgress,
	IconButton,
	List,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
} from '@mui/material';
import { ListPlus, Music, Play } from 'lucide-react';
import { useMusic } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import AddToPlaylistMenu from '@/components/music/AddToPlaylistMenu';
import { useState } from 'react';

const AllTracksView = () => {
	const { music, hasNextPage, isFetchingNextPage, lastItemRef } = useMusic();
	const { getMusicTitle, musicMetadata, getMusicArtist, addToQueue } = useGlobalMusic();
	const [menuAnchor, setMenuAnchor] = useState<{ el: HTMLElement; fileId: number } | null>(null);

	return (
		<>
			<List sx={{ width: '100%' }}>
				{music.map((item, index) => {
					const isLastItem = index === music.length - 1;
					return (
						<ListItem key={item.id} ref={isLastItem ? lastItemRef : null} sx={{ px: 0 }}>
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
									onClick={(e) => {
										e.stopPropagation();
										setMenuAnchor({ el: e.currentTarget, fileId: item.id });
									}}
								>
									<ListPlus size={18} />
								</IconButton>
								<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }} aria-label={`play ${item.name}`}>
									<Play />
								</IconButton>
							</ListItemButton>
						</ListItem>
					);
				})}
			</List>
			<AddToPlaylistMenu
				fileId={menuAnchor?.fileId ?? 0}
				anchorEl={menuAnchor?.el ?? null}
				onClose={() => setMenuAnchor(null)}
			/>

			{isFetchingNextPage && (
				<div style={{ display: 'flex', justifyContent: 'center', padding: 20 }}>
					<CircularProgress size={40} />
				</div>
			)}

			{!hasNextPage && music.length > 0 && (
				<div style={{ textAlign: 'center', padding: 20, color: '#888', fontSize: 14 }}>All music loaded</div>
			)}
		</>
	);
};

export default AllTracksView;
