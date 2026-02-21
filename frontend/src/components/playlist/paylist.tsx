import { IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { useMusic } from '../hooks/musicProvider/musicProvider';
import styles from './paylist.module.css';
import { Music, Play } from 'lucide-react';

const Playlist = () => {
	const { playlist, playTrack, getMusicTitle } = useMusic();
	return (
		<div className={styles.playlist}>
			<List sx={{ width: '100%' }}>
				{playlist.map((item) => (
					<ListItem key={item.id} sx={{ px: 0 }}>
						<ListItemButton onClick={() => playTrack(item)}>
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
