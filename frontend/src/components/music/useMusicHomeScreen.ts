import { getMusicRoute } from '@/app/routes';
import { musicNavigationItems } from '@/components/music/navigation';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { useMusic } from '@/components/providers/musicProvider/musicProvider';

export const useMusicHomeScreen = () => {
	const { status, music } = useMusic();
	const { currentTrack, getMusicArtist, getMusicTitle, hasQueue, queue } = useGlobalMusic();

	return {
		status,
		totalTracks: music.length,
		hasQueue,
		queueCount: queue.length,
		currentTrackTitle: currentTrack ? getMusicTitle(currentTrack) : '',
		currentTrackArtist: currentTrack ? getMusicArtist(currentTrack) : '',
		sections: musicNavigationItems
			.filter((item) => item.key !== 'home')
			.map((item) => ({
				...item,
				href: getMusicRoute(item.key),
			})),
	};
};
