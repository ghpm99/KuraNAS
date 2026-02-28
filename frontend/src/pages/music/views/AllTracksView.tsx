import {
	CircularProgress,
	IconButton,
	List,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
} from '@mui/material';
import { Music, Play } from 'lucide-react';
import { useMusic } from '@/components/hooks/musicProvider/musicProvider';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

const AllTracksView = () => {
	const { music, hasNextPage, isFetchingNextPage, lastItemRef } = useMusic();
	const { getMusicTitle, musicMetadata, getMusicArtist, addToQueue } = useGlobalMusic();

	return (
		<>
			<List sx={{ width: '100%' }}>
				{music.map((item, index) => {
					const isLastItem = index === music.length - 1;
					return (
						<ListItem key={item.id} ref={isLastItem ? lastItemRef : null} sx={{ px: 0 }}>
							<ListItemButton onClick={() => playTrack(item)}>
								<ListItemIcon>
									<Music />
								</ListItemIcon>
								<ListItemText
									primary={getMusicTitle(item)}
									secondary={`${getMusicArtist(item)} - ${musicMetadata(item)}`}
								/>
								<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }} aria-label={`play ${item.name}`}>
									<Play />
								</IconButton>
							</ListItemButton>
						</ListItem>
					);
				})}
			</List>

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
