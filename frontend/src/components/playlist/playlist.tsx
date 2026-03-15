import { IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { useGlobalMusic } from '../providers/GlobalMusicProvider';
import { getMusicTitle } from '@/utils/music';
import styles from './playlist.module.css';
import { Music, Play } from 'lucide-react';

const Playlist = () => {
	const { queue, playTrackFromQueue } = useGlobalMusic();
	return (
		<div className={styles.playlist}>
			<List sx={{ width: '100%' }}>
				{queue.map((item, index) => (
					<ListItem key={item.id} sx={{ px: 0 }}>
						<ListItemButton onClick={() => playTrackFromQueue(index)}>
							<ListItemIcon>
								<Music />
							</ListItemIcon>
							<ListItemText style={{ whiteSpace: 'nowrap' }} primary={getMusicTitle(item)} />
							<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }} aria-label={`play ${item.name}`}>
								<Play />
							</IconButton>
						</ListItemButton>
					</ListItem>
				))}
			</List>
		</div>
	);
};

export default Playlist;
