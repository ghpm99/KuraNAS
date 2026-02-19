import { formatSize } from '@/utils';
import { useMusic } from '../hooks/musicProvider/musicProvider';
import { useMusicPlayer } from '../hooks/musicPlayerProvider/musicPlayerProvider';
import { useIntersectionObserver } from '../hooks/IntersectionObserver/useIntersectionObserver';
import './musicContent.css';
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
import PlayerControl from '../playerControl/playerControl';

const MusicContent = () => {
	const { music, fetchNextPage, hasNextPage, isFetchingNextPage } = useMusic();
	const { playTrack } = useMusicPlayer();
	const { ref: lastItemRef } = useIntersectionObserver<HTMLLIElement>({
		enabled: hasNextPage && !isFetchingNextPage,
		rootMargin: '400px',
		onIntersect: () => {
			if (hasNextPage && !isFetchingNextPage) {
				fetchNextPage();
			}
		},
	});

	const musicMetadata = (music: { format: string; size: number; metadata?: any }): string => {
		const format = music.format ? `${music.format} - ` : '';
		const fileSize = formatSize(music.size);
		const duration = music.metadata?.duration ? formatDuration(music.metadata.duration) : '';
		return `${format}${fileSize}${duration ? ` - ${duration}` : ''}`;
	};

	const formatDuration = (seconds: number): string => {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	};

	const getMusicTitle = (music: any): string => {
		if (music.metadata?.title) {
			return music.metadata.title;
		}
		return music.name;
	};

	const getMusicArtist = (music: any): string => {
		if (music.metadata?.artist) {
			return music.metadata.artist;
		}
		return 'Unknown Artist';
	};

	return (
		<div className='file-content'>
			<PlayerControl />
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
				<div className='loading-indicator'>
					<CircularProgress size={40} />
				</div>
			)}

			{!hasNextPage && music.length > 0 && <div className='end-message'>All music loaded</div>}
		</div>
	);
};

export default MusicContent;
