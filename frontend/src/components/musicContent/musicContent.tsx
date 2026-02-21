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

import { useMusic } from '../hooks/musicProvider/musicProvider';
import PlayerControl from '../playerControl/playerControl';
import './musicContent.css';
import Playlist from '../playlist/paylist';

const MusicContent = () => {
	const {
		music,
		getMusicTitle,
		musicMetadata,
		hasNextPage,
		isFetchingNextPage,
		getMusicArtist,
		lastItemRef,
		playTrack,
		hasTrackInPlaylist,
	} = useMusic();

	return (
		<>
			<div className='music-content'>
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
				<PlayerControl />

				{isFetchingNextPage && (
					<div className='loading-indicator'>
						<CircularProgress size={40} />
					</div>
				)}

				{!hasNextPage && music.length > 0 && <div className='end-message'>All music loaded</div>}
			</div>
			{hasTrackInPlaylist && <Playlist />}
		</>
	);
};

export default MusicContent;
